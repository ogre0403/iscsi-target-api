package tgt

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	log "github.com/golang/glog"
	"github.com/ogre0403/iscsi-target-api/pkg/cfg"
	"net/http"
	"os"
	"os/exec"
	"regexp"
	"strings"
)

const (
	TGTADMIN    = "tgt-admin"
	TGTIMG      = "tgtimg"
	TGTSETUPLUN = "tgt-setup-lun"
	TARGETCONF  = "/etc/tgt/conf.d/iscsi-target-api.conf"
	BASEIMGPATH = "/var/lib/iscsi/"
)

type tgtd struct {
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

// todo: support LVM based volume
func (t *tgtd) CreateVolume(cfg *cfg.VolumeCfg) error {

	r, _ := regexp.Compile("[0-9]+m$")
	if !r.MatchString(cfg.Size) {
		return errors.New("size is media size (in megabytes), eg. 1024m")
	}

	fullImgPath := t.BaseImagePath + "/" + cfg.Path + "/" + cfg.Name

	if _, err := os.Stat(fullImgPath); !os.IsNotExist(err) {
		return errors.New(fmt.Sprintf("Image %s alreay exist", fullImgPath))
	}

	var stdout, stderr bytes.Buffer

	cmd := exec.Command("/bin/sh", "-c",
		fmt.Sprintf("%s --op new --device-type disk --type disk --size %s --file %s ", t.tgtimgCmd, cfg.Size, fullImgPath),
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

	if queryTargetId(cfg.TargetIQN) != "-1" {
		return errors.New(fmt.Sprintf("target %s already exist", cfg.TargetIQN))
	}

	volPath := t.BaseImagePath + "/" + cfg.Volume.Path + "/" + cfg.Volume.Name

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

	fullImgPath := t.BaseImagePath + "/" + cfg.Path + "/" + cfg.Name
	if err := os.Remove(fullImgPath); err != nil {
		return err
	}
	return nil
}

func (t *tgtd) Reload() error {
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

	c.JSON(http.StatusOK, cfg.Response{
		Error: false,
	})
}

func (t *tgtd) AttachLunAPI(c *gin.Context) {
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

	if err := t.Reload(); err != nil {
		respondWithError(c, http.StatusInternalServerError, "Reload tgtd config file fail: %s", err.Error())
		return
	}

	c.JSON(http.StatusOK, cfg.Response{
		Error: false,
	})
}

func (t *tgtd) DeleteVolumeAPI(c *gin.Context) {

}

func (t *tgtd) DeleteTargetAPI(c *gin.Context) {}

func respondWithError(c *gin.Context, code int, format string, args ...interface{}) {
	log.Errorf(format, args...)
	resp := cfg.Response{
		Error:   true,
		Message: fmt.Sprintf(format, args...),
	}
	c.JSON(code, resp)
	c.Abort()
}
