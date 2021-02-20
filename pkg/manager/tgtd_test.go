package manager

import (
	"errors"
	"flag"
	"github.com/ogre0403/go-lvm"
	"github.com/ogre0403/iscsi-target-api/pkg/model"
	"github.com/ogre0403/iscsi-target-api/pkg/target"
	"github.com/ogre0403/iscsi-target-api/pkg/volume"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

var mgr tgtd

func init() {

	mgr = tgtd{
		baseImagePath: BASEIMGPATH,
		tgt_adminCmd:  TGT_ADMIN,
		tgtimgCmd:     TGTIMG,
		tgtadmCmd:     TGTADM,
		targetConf:    "/tmp/iscsi-target-api.conf",
		chap:          &model.CHAP{},
		thinPool:      "pool0",
	}
	lvm.Initialize()
}

func TestNewTarget(t *testing.T) {
	flag.Parse()

	_, err := newTgtdMgr(&model.ManagerCfg{
		BaseImagePath: BASEIMGPATH,
		TargetConf:    TARGETCONF,
	})

	assert.NoError(t, err)
}

func TestTgtd_NewVolume(t *testing.T) {

	bv := volume.BasicVolume{
		Type:          volume.VolumeTypeTGTIMG,
		Group:         "test",
		Name:          "test",
		Size:          12,
		Unit:          lvm.MiB,
		ThinProvision: false,
	}

	_, err := mgr.NewVolume(&bv)

	assert.NoError(t, err)
}

func TestTgtd_AttachLun(t *testing.T) {
	bv := volume.BasicVolume{
		Type:          volume.VolumeTypeTGTIMG,
		Group:         "test",
		Name:          "test",
		Size:          20,
		Unit:          lvm.MiB,
		ThinProvision: false,
	}

	lc := &model.Lun{
		TargetIQN: "iqn.2017-07.com.hiroom2:ogre",
		Volume:    &bv,
	}

	err := mgr.CreateVolume(&bv)
	assert.NoError(t, err)

	t.Cleanup(func() {
		mgr.DeleteVolume(&bv)
	})

	err = mgr.AttachLun(lc)
	assert.NoError(t, err)

	err = mgr.AttachLun(lc)
	if assert.Error(t, err) {
		assert.Equal(t, errors.New("tgtadm: this target already exists"), err)
	}

	t.Cleanup(func() {
		err := mgr.DeleteTarget(&target.BasicTarget{
			TargetIQN: lc.TargetIQN,
		})
		assert.NoError(t, err)
	})

	err = mgr.Save()
	assert.NoError(t, err)

	t.Cleanup(func() {
		os.Remove(mgr.targetConf)
	})

}

func TestTgtd_AttachLVMLun(t *testing.T) {
	bv := volume.BasicVolume{
		Type:          volume.VolumeTypeLVM,
		Group:         "vg-0",
		Name:          "test",
		Size:          20,
		Unit:          lvm.MiB,
		ThinProvision: false,
	}

	lc := &model.Lun{
		TargetIQN: "iqn.2017-07.com.hiroom2:ogre",
		Volume:    &bv,
	}

	err := mgr.CreateVolume(&bv)
	assert.NoError(t, err)

	t.Cleanup(func() {
		mgr.DeleteVolume(&bv)
	})

	err = mgr.AttachLun(lc)
	assert.NoError(t, err)

	err = mgr.AttachLun(lc)
	if assert.Error(t, err) {
		assert.Equal(t, errors.New("tgtadm: this target already exists"), err)
	}

	t.Cleanup(func() {
		err := mgr.DeleteTarget(&target.BasicTarget{
			TargetIQN: lc.TargetIQN,
		})
		assert.NoError(t, err)
	})

	err = mgr.Save()
	assert.NoError(t, err)

	t.Cleanup(func() {
		os.Remove(mgr.targetConf)
	})

}
