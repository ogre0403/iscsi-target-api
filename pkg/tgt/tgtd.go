package tgt

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	log "github.com/golang/glog"
	"github.com/ogre0403/go-lvm"
	"github.com/ogre0403/iscsi-target-api/pkg/cfg"
	"io"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"sync/atomic"
)

const (
	TGTADMIN    = "tgt-admin"
	TGTIMG      = "tgtimg"
	LVM         = "lvm"
	TGTSETUPLUN = "tgt-setup-lun"
	TARGETCONF  = "/etc/tgt/conf.d/iscsi-target-api.conf"
	BASEIMGPATH = "/var/lib/iscsi/"
)

type tgtd struct {
	locker         uint32
	targetConf     string
	BaseImagePath  string
	tgtimgCmd      string
	tgtadminCmd    string
	tgtsetuplunCmd string
}

func newTgtdTarget(mgrCfg *cfg.ManagerCfg) (TargetManager, error) {

	t := &tgtd{
		BaseImagePath: mgrCfg.BaseImagePath,
		targetConf:    mgrCfg.TargetConf,
	}

	exist, e := isCmdExist(t)
	if exist == false {
		return nil, e
	}

	log.Info(fmt.Sprintf("found %s in %s", TGTADMIN, t.tgtadminCmd))
	log.Info(fmt.Sprintf("found %s in %s", TGTIMG, t.tgtimgCmd))
	log.Info(fmt.Sprintf("found %s in %s", TGTSETUPLUN, t.tgtsetuplunCmd))

	return t, nil

}

func isCmdExist(t *tgtd) (bool, error) {

	var stdout bytes.Buffer
	cmd := exec.Command("/bin/sh", "-c", "command -v "+TGTADMIN)
	cmd.Stdout = &stdout

	if err := cmd.Run(); err != nil {
		return false, errors.New(fmt.Sprintf("%s not found", TGTADMIN))
	}
	t.tgtadminCmd = strings.TrimSpace(string(stdout.Bytes()))

	var stdout1 bytes.Buffer
	cmd = exec.Command("/bin/sh", "-c", "command -v "+TGTIMG)
	cmd.Stdout = &stdout1
	if err := cmd.Run(); err != nil {
		return false, errors.New(fmt.Sprintf("%s not found", TGTIMG))
	}
	t.tgtimgCmd = strings.TrimSpace(string(stdout1.Bytes()))

	var stdout2 bytes.Buffer
	cmd = exec.Command("/bin/sh", "-c", "command -v "+TGTSETUPLUN)
	cmd.Stdout = &stdout2
	if err := cmd.Run(); err != nil {
		return false, errors.New(fmt.Sprintf("%s not found", TGTSETUPLUN))
	}
	t.tgtsetuplunCmd = strings.TrimSpace(string(stdout2.Bytes()))

	return true, nil
}

func (t *tgtd) CreateVolume(cfg *cfg.VolumeCfg) error {

	switch cfg.Type {
	case TGTIMG:
		log.V(2).Infof("Provision volume by %s ", TGTIMG)
		return tgtimgPovision(t, cfg)
	case LVM:
		log.V(2).Infof("Provision volume by %s ", LVM)
		return lvmProvision(cfg)
	default:
		log.Errorf("%s is not supported volume provision tool", cfg.Type)
		return errors.New(fmt.Sprintf("%s is not supported volume provision tool", cfg.Type))
	}
}

// todo: refact to volume
func lvmProvision(cfg *cfg.VolumeCfg) error {

	vgo, err := lvm.VgOpen(cfg.Group, "w")

	if err != nil {
		return err
	}

	defer vgo.Close()

	_, err = vgo.CreateLvLinear(cfg.Name, uint64(lvm.UnitTranslate(cfg.Size, cfg.Unit, lvm.B)))

	if err != nil {
		return err
	}
	return nil
}

// todo: support LVM resize
func lvmResize(cfg *cfg.VolumeCfg) error {
	return errors.New("not implemented error")
}

func tgtimgPovision(t *tgtd, cfg *cfg.VolumeCfg) error {

	sizeUnit := fmt.Sprintf("%dm", uint64(lvm.UnitTranslate(cfg.Size, cfg.Unit, lvm.MiB)))

	fullImgPath := t.BaseImagePath + "/" + cfg.Group + "/" + cfg.Name

	if _, err := os.Stat(fullImgPath); !os.IsNotExist(err) {
		return errors.New(fmt.Sprintf("Image %s alreay exist", fullImgPath))
	}

	if _, err := os.Stat(t.BaseImagePath + "/" + cfg.Group); os.IsNotExist(err) {
		if err := os.MkdirAll(t.BaseImagePath+"/"+cfg.Group, 0755); err != nil {
			return errors.New("unable to create directory to provision new volume: " + err.Error())
		}
	}

	var stdout, stderr bytes.Buffer

	cmd := exec.Command("/bin/sh", "-c",
		fmt.Sprintf("%s --op new --device-type disk --type disk --size %s --file %s ", t.tgtimgCmd, sizeUnit, fullImgPath),
	)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return errors.New(fmt.Sprintf(string(stderr.Bytes())))
	}
	log.Info(string(stdout.Bytes()))
	return nil
}

func (t *tgtd) AttachLun(cfg *cfg.LunCfg) error {

	var volPath string

	switch cfg.Volume.Type {
	case TGTIMG:
		volPath = t.BaseImagePath + "/" + cfg.Volume.Group + "/" + cfg.Volume.Name
	case LVM:
		volPath = "/dev/" + cfg.Volume.Group + "/" + cfg.Volume.Name
	default:
		log.Errorf("%s is not supported volume provision tool", cfg.Volume.Type)
		return errors.New(fmt.Sprintf("%s is not supported volume provision tool", cfg.Volume.Type))
	}

	if queryTargetId(cfg.TargetIQN) != "-1" {
		return errors.New(fmt.Sprintf("target %s already exist", cfg.TargetIQN))
	}

	var stdout, stderr bytes.Buffer
	cmd := exec.Command("/bin/sh", "-c",
		fmt.Sprintf("%s  -n %s -d %s ", t.tgtsetuplunCmd, cfg.TargetIQN, volPath),
	)

	log.Info(cmd.String())

	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return errors.New(fmt.Sprintf(string(stderr.Bytes())))
	}

	log.Info(string(stdout.Bytes()))

	return nil
}

func (t *tgtd) DeleteTarget(cfg *cfg.TargetCfg) error {

	tid := queryTargetId(cfg.TargetIQN)
	if tid == "-1" {
		return errors.New(fmt.Sprintf("target %s not found", cfg.TargetIQN))
	}
	var stdout, stderr bytes.Buffer
	cmd := exec.Command("/bin/sh", "-c",
		fmt.Sprintf("%s --delete tid=%s", t.tgtadminCmd, tid),
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

	switch cfg.Type {
	case TGTIMG:
		log.V(2).Infof("Delete volume by %s ", TGTIMG)
		return tgtimgDelete(t, cfg)
	case LVM:
		log.V(2).Infof("Delete volume by %s ", LVM)
		return lvmDelete(cfg)
	default:
		log.Errorf("%s is not supported volume provision tool", cfg.Type)
		return errors.New(fmt.Sprintf("%s is not supported volume provision tool", cfg.Type))
	}
}

// todo: refact to volume
func tgtimgDelete(t *tgtd, cfg *cfg.VolumeCfg) error {

	volSubDir := t.BaseImagePath + "/" + cfg.Group
	fullImgPath := volSubDir + "/" + cfg.Name
	if err := os.Remove(fullImgPath); err != nil {
		return err
	}

	// remove dir when there is no volume
	if isEmpty, _ := isDirEmpty(volSubDir); isEmpty {
		if err := os.Remove(volSubDir); err != nil {
			return err
		}
	}
	return nil

}

func lvmDelete(cfg *cfg.VolumeCfg) error {

	vgo, err := lvm.VgOpen(cfg.Group, "w")
	if err != nil {
		return err
	}
	defer vgo.Close()

	lv, err := vgo.LvFromName(cfg.Name)
	if err != nil {
		return err
	}

	// Remove LV
	err = lv.Remove()
	if err != nil {
		return err
	}
	return nil
}

func (t *tgtd) Save() error {
	var stdout, stderr bytes.Buffer

	cmd := exec.Command("/bin/sh", "-c",
		fmt.Sprintf("%s --dump > %s ", t.tgtadminCmd, t.targetConf),
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

func isDirEmpty(name string) (bool, error) {
	f, err := os.Open(name)
	if err != nil {
		return false, err
	}
	defer f.Close()

	_, err = f.Readdirnames(1) // Or f.Readdir(1)
	if err == io.EOF {
		return true, nil
	}
	return false, err // Either not empty or error, suits both cases
}
