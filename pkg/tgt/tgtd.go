package tgt

import (
	"bytes"
	"errors"
	"fmt"
	log "github.com/golang/glog"
	"github.com/ogre0403/iscsi-target-api/pkg/cfg"
	"github.com/ogre0403/iscsi-target-api/pkg/volume"
	"net"
	"os/exec"
	"strconv"
	"strings"
)

const (
	TGT_ADMIN   = "tgt-admin"
	TGTADM      = "tgtadm"
	TGTIMG      = "tgtimg"
	TARGETCONF  = "/etc/tgt/conf.d/iscsi-target-api.conf"
	BASEIMGPATH = "/var/lib/iscsi/"
)

type tgtd struct {
	chap          *cfg.CHAP
	targetConf    string
	baseImagePath string
	tgtimgCmd     string
	tgt_adminCmd  string
	tgtadmCmd     string
	thinPool      string
}

func newTgtdTarget(mgrCfg *cfg.ManagerCfg) (TargetManager, error) {

	t := &tgtd{
		baseImagePath: mgrCfg.BaseImagePath,
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

	log.V(3).Info(cmd.String())
	if err := cmd.Run(); err != nil {
		return false, errors.New(fmt.Sprintf("%s not found", TGT_ADMIN))
	}
	t.tgt_adminCmd = strings.TrimSpace(string(stdout.Bytes()))

	var stdout1 bytes.Buffer
	cmd = exec.Command("/bin/sh", "-c", "command -v "+TGTIMG)
	cmd.Stdout = &stdout1
	log.V(3).Info(cmd.String())
	if err := cmd.Run(); err != nil {
		return false, errors.New(fmt.Sprintf("%s not found", TGTIMG))
	}
	t.tgtimgCmd = strings.TrimSpace(string(stdout1.Bytes()))

	var stdout2 bytes.Buffer
	cmd = exec.Command("/bin/sh", "-c", "command -v "+TGTADM)
	cmd.Stdout = &stdout2
	log.V(3).Info(cmd.String())
	if err := cmd.Run(); err != nil {
		return false, errors.New(fmt.Sprintf("%s not found", TGTADM))
	}
	t.tgtadmCmd = strings.TrimSpace(string(stdout2.Bytes()))

	return true, nil
}

func (t *tgtd) CreateVolume(vol *volume.BasicVolume) error {

	actualVol, err := t.NewVolume(vol)
	if err != nil {
		return err
	}

	return actualVol.Create()
}

func (t *tgtd) AttachLun(lun *cfg.LunCfg) error {

	actualVol, err := t.NewVolume(lun.Volume)
	if err != nil {
		return err
	}

	volPath, err := actualVol.Path()
	if err != nil {
		return err
	}

	if _, err := actualVol.IsExist(); err != nil {
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
		log.Infof("target %s create fail because LUN %s can not be added, rollback target creation",
			target.TargetIQN, volPath)
		t.DeleteTarget(&target)
		return err
	}

	if err := target.SetACL(lun.AclIpList); err != nil {
		log.Infof("target %s create fail because ACL can not be added, rollback target creation",
			target.TargetIQN)
		t.DeleteTarget(&target)
		return err
	}

	if lun.EnableChap {
		if err := target.AddCHAP(t.chap); err != nil {
			log.Infof("target %s create fail because CHAP can not be added, rollback target creation",
				target.TargetIQN)
			t.DeleteTarget(&target)
			return err
		}
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

func (t *tgtd) DeleteVolume(vol *volume.BasicVolume) error {
	actualVol, err := t.NewVolume(vol)
	if err != nil {
		return err
	}

	return actualVol.Delete()
}

func (t *tgtd) Save() error {
	var stdout, stderr bytes.Buffer

	// tgt-admin --dump | \
	// sed -r \
	// -e 's/(incominguser.*)(PLEASE_CORRECT_THE_PASSWORD)/\1CXC/g' \
	// -e 's/(outgoinguser.*)(PLEASE_CORRECT_THE_PASSWORD)/\1Cxxx/g' \
	// > /etc/tgt/conf.d/iscsi-target-api.conf
	cmd := exec.Command("/bin/sh", "-c",
		fmt.Sprintf("%s --dump |"+
			" sed -r "+
			" -e 's/(incominguser.*)(PLEASE_CORRECT_THE_PASSWORD)/\\1%s/g' "+
			" -e 's/(outgoinguser.*)(PLEASE_CORRECT_THE_PASSWORD)/\\1%s/g' "+
			" > %s ",
			t.tgt_adminCmd, t.chap.CHAPPassword, t.chap.CHAPPasswordIn, t.targetConf),
	)

	log.V(3).Info(cmd.String())

	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return errors.New(fmt.Sprintf(string(stderr.Bytes())))
	}

	return nil
}

func (t *tgtd) NewVolume(basic *volume.BasicVolume) (volume.Volume, error) {
	switch basic.Type {
	case volume.VolumeTypeLVM:
		log.Infof("Initialize %s volume", volume.VolumeTypeLVM)
		return &volume.LvmVolume{
			BasicVolume: *basic,
			ThinPool:    t.thinPool,
		}, nil
	case volume.VolumeTypeTGTIMG:
		log.Infof("Initialize %s volume", volume.VolumeTypeTGTIMG)

		return &volume.ImageVolume{
			BasicVolume:   *basic,
			BaseImagePath: t.baseImagePath,
			TgtimgCmd:     t.tgtimgCmd,
		}, nil
	default:
		log.Infof("%s is not supported volume type", basic.Type)
		return nil, errors.New(fmt.Sprintf("%s is not supported volume type", basic.Type))
	}
}
