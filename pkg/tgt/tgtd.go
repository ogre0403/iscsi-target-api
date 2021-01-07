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
		targetConf:    TARGETCONF,
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

	// todo: check target exist
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

// Deprecated:
func (t *tgtd) CreateTarget(cfg *cfg.TargetCfg) error {

	return errors.New("not supported")

	if !validateIQN(cfg.TargetIQN) {
		return errors.New(fmt.Sprintf("%s is not valid IQN format", cfg.TargetIQN))
	}

	tid := cfg.TargetId
	if cfg.TargetId == "" {
		tid = queryMaxTargetId()
		if tid == "-1" {
			return errors.New("Can not find maximun target id")
		}
	}

	var stdout, stderr bytes.Buffer

	cmd := exec.Command("/bin/sh", "-c",
		fmt.Sprintf("%s --lld iscsi --op new --mode target --tid %s -T %s ", t.tgtadminCmd, tid, cfg.TargetIQN),
	)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return errors.New(fmt.Sprintf(string(stderr.Bytes())))
	}

	log.Info(fmt.Sprintf("Created Target %s with tid %s", cfg.TargetIQN, tid))
	return nil
}
