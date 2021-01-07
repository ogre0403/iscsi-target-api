package tgt

import (
	"errors"
	"flag"
	"fmt"
	"github.com/ogre0403/iscsi-target-api/pkg/cfg"
	"github.com/stretchr/testify/assert"
	"os"
	"os/exec"
	"testing"
)

var mgr tgtd

func init() {

	mgr = tgtd{
		BaseImagePath:  BASEIMGPATH,
		tgtadminCmd:    TGTADMIN,
		tgtimgCmd:      TGTIMG,
		tgtsetuplunCmd: TGTSETUPLUN,
		targetConf:     "/tmp/iscsi-target-api.conf",
	}
}

func TestNewTarget(t *testing.T) {
	flag.Parse()

	_, err := newTgtdTarget(&cfg.ManagerCfg{
		BaseImagePath: BASEIMGPATH,
		TargetConf:    TARGETCONF,
	})

	assert.NoError(t, err)
}

func TestTgtd_CreateVolume(t *testing.T) {

	vc := &cfg.VolumeCfg{
		Name: "test.img",
		Size: "100m",
	}
	err := mgr.CreateVolume(vc)
	assert.NoError(t, err)

	fullImgPath := mgr.BaseImagePath + "/" + vc.Path + "/" + vc.Name
	t.Cleanup(func() {
		os.Remove(fullImgPath)
	})

}

func TestTgtd_CreateTarget(t *testing.T) {
	tc := &cfg.TargetCfg{
		TargetId:  "1",
		TargetIQN: "iqn.2017-07.com.hiroom2:aaadd",
	}

	err := mgr.CreateTarget(tc)
	if assert.Error(t, err) {
		assert.Equal(t, errors.New("not supported"), err)
	}

	t.Cleanup(func() {
		cmd := exec.Command("/bin/sh", "-c",
			fmt.Sprintf("tgt-admin --delete tid=%s", tc.TargetId),
		)
		cmd.Run()
	})

}

func TestTgtd_AttachLun(t *testing.T) {
	vc := &cfg.VolumeCfg{
		Name: "test.img",
		Size: "100m",
	}

	lc := &cfg.LunCfg{
		TargetIQN: "iqn.2017-07.com.hiroom2:ogre",
		Volume:    vc,
	}

	err := mgr.CreateVolume(vc)
	assert.NoError(t, err)

	fullImgPath := mgr.BaseImagePath + "/" + vc.Path + "/" + vc.Name
	t.Cleanup(func() {
		os.Remove(fullImgPath)
	})

	err = mgr.AttachLun(lc)
	assert.NoError(t, err)

	err = mgr.AttachLun(lc)
	if assert.Error(t, err) {
		assert.Equal(t, errors.New(fmt.Sprintf("target %s already exist", lc.TargetIQN)), err)
	}

	tid := queryTargetId(lc.TargetIQN)
	t.Cleanup(func() {
		cmd := exec.Command("/bin/sh", "-c",
			fmt.Sprintf("tgt-admin --delete tid=%s", tid),
		)
		cmd.Run()
	})
}

func TestTgtd_Reload(t *testing.T) {
	vc := &cfg.VolumeCfg{
		Name: "test.img",
		Size: "100m",
	}

	lc := &cfg.LunCfg{
		TargetIQN: "iqn.2017-07.com.hiroom2:ogre",
		Volume:    vc,
	}

	err := mgr.CreateVolume(vc)
	assert.NoError(t, err)

	fullImgPath := mgr.BaseImagePath + "/" + vc.Path + "/" + vc.Name
	t.Cleanup(func() {
		os.Remove(fullImgPath)
	})

	err = mgr.AttachLun(lc)
	assert.NoError(t, err)

	tid := queryTargetId(lc.TargetIQN)
	t.Cleanup(func() {
		cmd := exec.Command("/bin/sh", "-c",
			fmt.Sprintf("tgt-admin --delete tid=%s", tid),
		)
		cmd.Run()
	})

	err = mgr.Reload()
	assert.NoError(t, err)

	t.Cleanup(func() {
		os.Remove(mgr.targetConf)
	})

}
