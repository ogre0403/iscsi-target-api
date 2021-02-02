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

func (t *tgtd) AttachLun(cfg *cfg.LunCfg) error {

	t.setupVol(cfg.Volume)
	volPath, err := cfg.Volume.Path()
	if err != nil {
		return err
	}

	if _, err := cfg.Volume.IsExist(); err != nil {
		return err
	}

	for _, ip := range cfg.AclIpList {
		_, _, e := net.ParseCIDR(ip)
		if e != nil && net.ParseIP(ip) == nil {
			return errors.New(fmt.Sprintf("%s is invalid ip format ", ip))
		}
	}

	if queryTargetId(cfg.TargetIQN) != "-1" {
		return errors.New(fmt.Sprintf("target %s already exist", cfg.TargetIQN))
	}

	tid := queryMaxTargetId()
	var stdout, stderr bytes.Buffer

	// create target
	// eg. tgtadm --lld iscsi --op new --mode target --tid 1 -T iqn.2017-07.com.hiroom2:debian-9
	cmd := exec.Command("/bin/sh", "-c",
		fmt.Sprintf("%s --lld iscsi --op new --mode target --tid %d -T %s ", t.tgtadmCmd, tid+1, cfg.TargetIQN),
	)
	log.Info(cmd.String())

	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return errors.New(fmt.Sprintf(string(stderr.Bytes())))
	}
	log.Info(string(stdout.Bytes()))

	// setup LUN
	// eg. tgtadm --lld iscsi --op new --mode logicalunit --tid 1 --lun 1 -b /var/lib/iscsi/10m-$i.img
	cmd = exec.Command("/bin/sh", "-c",
		fmt.Sprintf("%s --lld iscsi --op new --mode logicalunit --tid %d --lun 1 -b %s ", t.tgtadmCmd, tid+1, volPath),
	)
	log.Info(cmd.String())

	stderr.Reset()
	stdout.Reset()
	if err := cmd.Run(); err != nil {
		return errors.New(fmt.Sprintf(string(stderr.Bytes())))
	}
	log.Info(string(stdout.Bytes()))

	if len(cfg.AclIpList) == 0 {
		cmd = exec.Command("/bin/sh", "-c",
			fmt.Sprintf("%s --lld iscsi --op bind --mode target --tid %d -I %s ", t.tgtadmCmd, tid+1, "ALL"),
		)
		log.Info(cmd.String())

		stderr.Reset()
		stdout.Reset()
		if err := cmd.Run(); err != nil {
			return errors.New(fmt.Sprintf(string(stderr.Bytes())))
		}
		log.Info(string(stdout.Bytes()))
		return nil
	}

	for _, ip := range cfg.AclIpList {
		cmd = exec.Command("/bin/sh", "-c",
			fmt.Sprintf("%s --lld iscsi --op bind --mode target --tid %d -I %s ", t.tgtadmCmd, tid+1, ip),
		)
		log.Info(cmd.String())

		stderr.Reset()
		stdout.Reset()
		if err := cmd.Run(); err != nil {
			return errors.New(fmt.Sprintf(string(stderr.Bytes())))
		}
		log.Info(string(stdout.Bytes()))
	}

	return nil
}

func (t *tgtd) DeleteTarget(cfg *cfg.TargetCfg) error {

	tid := queryTargetId(cfg.TargetIQN)
	if tid == "-1" {
		return errors.New(fmt.Sprintf("target %s not found", cfg.TargetIQN))
	}
	var stdout, stderr bytes.Buffer
	cmd := exec.Command("/bin/sh", "-c",
		fmt.Sprintf("%s --delete tid=%s", t.tgt_adminCmd, tid),
	)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return errors.New(fmt.Sprintf(string(stderr.Bytes())))
	}

	log.Info(string(stdout.Bytes()))
	return nil
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
