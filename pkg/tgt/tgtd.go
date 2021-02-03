package tgt

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	log "github.com/golang/glog"
	"github.com/ogre0403/iscsi-target-api/pkg/cfg"
	"net"
	"net/http"
	"os/exec"
	"strconv"
	"strings"
	"sync/atomic"
)

const (
	TGT_ADMIN   = "tgt-admin"
	TGTADM      = "tgtadm"
	TGTIMG      = "tgtimg"
	LVM         = "lvm"
	TARGETCONF  = "/etc/tgt/conf.d/iscsi-target-api.conf"
	BASEIMGPATH = "/var/lib/iscsi/"
)

type tgtd struct {
	chap          *cfg.CHAP
	locker        uint32
	targetConf    string
	BaseImagePath string
	tgtimgCmd     string
	tgt_adminCmd  string
	tgtadmCmd     string
	thinPool      string
}

func newTgtdTarget(mgrCfg *cfg.ManagerCfg) (TargetManager, error) {

	t := &tgtd{
		BaseImagePath: mgrCfg.BaseImagePath,
		targetConf:    mgrCfg.TargetConf,
		thinPool:      mgrCfg.ThinPool,
	}

	t.chap = mgrCfg.CHAP

	exist, e := isCmdExist(t)
	if exist == false {
		return nil, e
	}

	log.Info(fmt.Sprintf("found %s in %s", TGT_ADMIN, t.tgt_adminCmd))
	log.Info(fmt.Sprintf("found %s in %s", TGTIMG, t.tgtimgCmd))
	log.Info(fmt.Sprintf("found %s in %s", TGTADM, t.tgtadmCmd))

	return t, nil

}

func isCmdExist(t *tgtd) (bool, error) {

	var stdout bytes.Buffer
	cmd := exec.Command("/bin/sh", "-c", "command -v "+TGT_ADMIN)
	cmd.Stdout = &stdout

	if err := cmd.Run(); err != nil {
		return false, errors.New(fmt.Sprintf("%s not found", TGT_ADMIN))
	}
	t.tgt_adminCmd = strings.TrimSpace(string(stdout.Bytes()))

	var stdout1 bytes.Buffer
	cmd = exec.Command("/bin/sh", "-c", "command -v "+TGTIMG)
	cmd.Stdout = &stdout1
	if err := cmd.Run(); err != nil {
		return false, errors.New(fmt.Sprintf("%s not found", TGTIMG))
	}
	t.tgtimgCmd = strings.TrimSpace(string(stdout1.Bytes()))

	var stdout2 bytes.Buffer
	cmd = exec.Command("/bin/sh", "-c", "command -v "+TGTADM)
	cmd.Stdout = &stdout2
	if err := cmd.Run(); err != nil {
		return false, errors.New(fmt.Sprintf("%s not found", TGTADM))
	}
	t.tgtadmCmd = strings.TrimSpace(string(stdout2.Bytes()))

	return true, nil
}

func (t *tgtd) CreateVolume(cfg *cfg.VolumeCfg) error {
	t.setupVol(cfg)
	return cfg.Create()
}

func (t *tgtd) AttachLun(lun *cfg.LunCfg) error {

	t.setupVol(lun.Volume)
	volPath, err := lun.Volume.Path()
	if err != nil {
		return err
	}

	if _, err := lun.Volume.IsExist(); err != nil {
		return err
	}

	for _, ip := range lun.AclIpList {
		_, _, e := net.ParseCIDR(ip)
		if e != nil && net.ParseIP(ip) == nil {
			return errors.New(fmt.Sprintf("%s is invalid ip format ", ip))
		}
	}

	if queryTargetId(lun.TargetIQN) != "-1" {
		return errors.New(fmt.Sprintf("target %s already exist", lun.TargetIQN))
	}

	tid := queryMaxTargetId()
	target := cfg.TargetCfg{
		TargetToolCli: t.tgtadmCmd,
		TargetId:      strconv.Itoa(tid + 1),
		TargetIQN:     lun.TargetIQN,
	}

	if err := target.Create(); err != nil {
		return err
	}

	if err := target.AddLun(volPath); err != nil {
		return err
	}

	if err := target.SetACL(lun.AclIpList); err != nil {
		return err
	}

	if err := target.AddCHAP(t.chap); err != nil {
		return err
	}

	return nil
}

func (t *tgtd) DeleteTarget(target *cfg.TargetCfg) error {

	tid := queryTargetId(target.TargetIQN)
	if tid == "-1" {
		return errors.New(fmt.Sprintf("target %s not found", target.TargetIQN))
	}

	tar := cfg.TargetCfg{
		TargetToolCli: t.tgt_adminCmd,
		TargetId:      tid,
		TargetIQN:     target.TargetIQN,
	}
	return tar.Delete()
}

func (t *tgtd) DeleteVolume(cfg *cfg.VolumeCfg) error {
	t.setupVol(cfg)
	return cfg.Delete()
}

func (t *tgtd) Save() error {
	var stdout, stderr bytes.Buffer

	cmd := exec.Command("/bin/sh", "-c",
		fmt.Sprintf("%s --dump > %s ", t.tgt_adminCmd, t.targetConf),
	)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return errors.New(fmt.Sprintf(string(stderr.Bytes())))
	}
	return nil
}

func (t *tgtd) CreateVolumeAPI(c *gin.Context) {

	if !atomic.CompareAndSwapUint32(&t.locker, 0, 1) {
		respondWithError(c, http.StatusTooManyRequests, "Another API is executing")
		return
	}
	defer atomic.StoreUint32(&t.locker, 0)

	var req cfg.VolumeCfg
	err := c.BindJSON(&req)

	if err != nil {
		respondWithError(c, http.StatusBadRequest, "Bind VolumeCfg fail: %s", err.Error())
		return
	}

	if err := t.CreateVolume(&req); err != nil {
		respondWithError(c, http.StatusInternalServerError, "Create Volume fail: %s", err.Error())
		return
	}

	responseWithOk(c)
}

func (t *tgtd) AttachLunAPI(c *gin.Context) {

	if !atomic.CompareAndSwapUint32(&t.locker, 0, 1) {
		respondWithError(c, http.StatusTooManyRequests, "Another API is executing")
		return
	}
	defer atomic.StoreUint32(&t.locker, 0)

	var req cfg.LunCfg
	err := c.BindJSON(&req)

	if err != nil {
		respondWithError(c, http.StatusBadRequest, "Bind LunCfg fail: %s", err.Error())
		return
	}

	if err := t.AttachLun(&req); err != nil {
		respondWithError(c, http.StatusInternalServerError, "Attach LUN fail: %s", err.Error())
		return
	}

	if err := t.Save(); err != nil {
		respondWithError(c, http.StatusInternalServerError, "Save tgtd config file fail: %s", err.Error())
		return
	}

	responseWithOk(c)
}

func (t *tgtd) DeleteVolumeAPI(c *gin.Context) {

	if !atomic.CompareAndSwapUint32(&t.locker, 0, 1) {
		respondWithError(c, http.StatusTooManyRequests, "Another API is executing")
		return
	}
	defer atomic.StoreUint32(&t.locker, 0)

	var req cfg.VolumeCfg
	err := c.BindJSON(&req)

	if err != nil {
		respondWithError(c, http.StatusBadRequest, "Bind VolumeCfg fail: %s", err.Error())
		return
	}

	if err := t.DeleteVolume(&req); err != nil {
		respondWithError(c, http.StatusInternalServerError, "Delete Volume fail: %s", err.Error())
		return
	}

	responseWithOk(c)
}

func (t *tgtd) DeleteTargetAPI(c *gin.Context) {

	if !atomic.CompareAndSwapUint32(&t.locker, 0, 1) {
		respondWithError(c, http.StatusTooManyRequests, "Another API is executing")
		return
	}
	defer atomic.StoreUint32(&t.locker, 0)

	var req cfg.TargetCfg
	err := c.BindJSON(&req)

	if err != nil {
		respondWithError(c, http.StatusBadRequest, "Bind TargetCfg fail: %s", err.Error())
		return
	}

	if err := t.DeleteTarget(&req); err != nil {
		respondWithError(c, http.StatusInternalServerError, "Delete Target fail: %s", err.Error())
		return
	}

	if err := t.Save(); err != nil {
		respondWithError(c, http.StatusInternalServerError, "Save tgtd config file fail: %s", err.Error())
		return
	}

	responseWithOk(c)
}

func (t *tgtd) setupVol(v *cfg.VolumeCfg) {
	v.SetBaseImgPath(t.BaseImagePath)
	v.SetTgtimgCmd(t.tgtimgCmd)
}

func respondWithError(c *gin.Context, code int, format string, args ...interface{}) {
	log.Errorf(format, args...)
	resp := cfg.Response{
		Error:   true,
		Message: fmt.Sprintf(format, args...),
	}
	c.JSON(code, resp)
	c.Abort()
}

func responseWithOk(c *gin.Context) {
	c.JSON(http.StatusOK, cfg.Response{
		Error:   false,
		Message: "Successful",
	})
}
