package tgt

import (
	"errors"
	"flag"
	"fmt"
	"github.com/ogre0403/iscsi-target-api/pkg/cfg"
	"github.com/stretchr/testify/assert"
	"os"
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
		Path: "p",
		Size: "100m",
	}
	err := mgr.CreateVolume(vc)
	assert.NoError(t, err)

	t.Cleanup(func() {
		err := mgr.DeleteVolume(vc)
		assert.NoError(t, err)
	})

}

func TestTgtd_AttachLun(t *testing.T) {
	vc := &cfg.VolumeCfg{
		Name: "test.img",
		Path: "p",
		Size: "100m",
	}

	lc := &cfg.LunCfg{
		TargetIQN: "iqn.2017-07.com.hiroom2:ogre",
		Volume:    vc,
	}

	err := mgr.CreateVolume(vc)
	assert.NoError(t, err)

	t.Cleanup(func() {
		mgr.DeleteVolume(vc)
	})

	err = mgr.AttachLun(lc)
	assert.NoError(t, err)

	err = mgr.AttachLun(lc)
	if assert.Error(t, err) {
		assert.Equal(t, errors.New(fmt.Sprintf("target %s already exist", lc.TargetIQN)), err)
	}

	t.Cleanup(func() {
		cfg := &cfg.TargetCfg{
			TargetIQN: lc.TargetIQN,
		}
		err := mgr.DeleteTarget(cfg)
		assert.NoError(t, err)
	})

}

func TestTgtd_Reload(t *testing.T) {
	vc := &cfg.VolumeCfg{
		Name: "test.img",
		Path: "p",
		Size: "100m",
	}

	lc := &cfg.LunCfg{
		TargetIQN: "iqn.2017-07.com.hiroom2:ogre",
		Volume:    vc,
	}

	err := mgr.CreateVolume(vc)
	assert.NoError(t, err)

	//fullImgPath := mgr.BaseImagePath + "/" + vc.Path + "/" + vc.Name
	t.Cleanup(func() {
		mgr.DeleteVolume(vc)
	})

	err = mgr.AttachLun(lc)
	assert.NoError(t, err)

	t.Cleanup(func() {
		cfg := &cfg.TargetCfg{
			TargetIQN: lc.TargetIQN,
		}
		err := mgr.DeleteTarget(cfg)
		assert.NoError(t, err)
	})

	err = mgr.Save()
	assert.NoError(t, err)

	t.Cleanup(func() {
		os.Remove(mgr.targetConf)
	})

}
