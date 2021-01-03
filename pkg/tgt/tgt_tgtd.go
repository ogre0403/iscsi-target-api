package tgt

import (
	"bytes"
	"errors"
	"fmt"
	log "github.com/golang/glog"
	"github.com/ogre0403/iscsi-target-api/pkg/cfg"
	"os"
	"os/exec"
	"strings"
)

const (
	TGTADM = "tgtadm"
	TGTIMG = "tgtimg"
)

type tgtd struct {
	BaseImagePath string
	tgtimgCmd     string
	tgtadmCmd     string
}

func newTgtdTarget(mgrCfg *cfg.ManagerCfg) (TargetManager, error) {

	exist, tgtadm, tgtimg, e := isCmdExist()
	if exist == false {
		return nil, e
	}

	log.Info(fmt.Sprintf("found %s in %s", TGTADM, tgtadm))
	log.Info(fmt.Sprintf("found %s in %s", TGTIMG, tgtimg))

	return &tgtd{
		BaseImagePath: mgrCfg.BaseImagePath,
		tgtadmCmd:     tgtadm,
		tgtimgCmd:     tgtimg,
	}, nil
}

func isCmdExist() (bool, string, string, error) {

	var stdout bytes.Buffer
	cmd := exec.Command("/bin/sh", "-c", "command -v "+TGTADM)
	cmd.Stdout = &stdout

	if err := cmd.Run(); err != nil {
		return false, "", "", errors.New(fmt.Sprintf("%s not found", TGTADM))
	}
	tgtadmCmd := strings.TrimSpace(string(stdout.Bytes()))

	var stdout1 bytes.Buffer
	cmd = exec.Command("/bin/sh", "-c", "command -v "+TGTIMG)
	cmd.Stdout = &stdout1
	if err := cmd.Run(); err != nil {
		return false, "", "", errors.New(fmt.Sprintf("%s not found", TGTIMG))
	}
	tgtimgCmd := strings.TrimSpace(string(stdout1.Bytes()))

	return true, tgtadmCmd, tgtimgCmd, nil
}

// todo: support LVM based volume
func (t *tgtd) CreateVolume(cfg *cfg.VolumeCfg) error {

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

func (t *tgtd) CreateTarget(cfg *cfg.TargetCfg) error {
	return nil
}
func (t *tgtd) CreateLun(cfg *cfg.LunCfg) error {
	return nil
}
func (t *tgtd) AttachLun() error {
	return nil
}
func (t *tgtd) Reload() error {
	return nil
}
