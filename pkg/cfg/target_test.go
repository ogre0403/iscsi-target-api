package cfg

import (
	"github.com/ogre0403/go-lvm"
	"github.com/stretchr/testify/assert"
	"testing"
)

var vol = VolumeCfg{
	Type:  VolumeTypeTGTIMG,
	Group: "test",
	Name:  "test1",
	Size:  12,
	Unit:  lvm.MiB,
}

func TestTargetCfg_Create(t *testing.T) {

	target := TargetCfg{
		TargetToolCli: "tgtadm",
		TargetId:      "99",
		TargetIQN:     "iqn.test.com.local:local",
	}

	e := target.Create()

	assert.NoError(t, e)

}

func TestTargetCfg_AddLun(t *testing.T) {

	vol.SetBaseImgPath(ImgPath)
	vol.SetTgtimgCmd(TgtimgCmd)
	e := vol.Create()
	assert.NoError(t, e)

	path, _ := vol.Path()

	target := TargetCfg{
		TargetToolCli: "tgtadm",
		TargetId:      "99",
		TargetIQN:     "iqn.test.com.local:local",
	}

	e = target.AddLun(path)
	assert.NoError(t, e)

	vol.Delete()
}

func TestTargetCfg_ACL(t *testing.T) {
	target := TargetCfg{
		TargetToolCli: "tgtadm",
		TargetId:      "99",
		TargetIQN:     "iqn.test.com.local:local",
	}

	e := target.SetACL([]string{})
	assert.NoError(t, e)
	e = target.SetACL([]string{"192.168.1.1", "192.168.1.2"})
	assert.NoError(t, e)

}

func TestTargetCfg_AddCHAP(t *testing.T) {
	target := TargetCfg{
		TargetToolCli: "tgtadm",
		TargetId:      "99",
		TargetIQN:     "iqn.test.com.local:local",
	}

	e := target.AddCHAP(&CHAP{})
	assert.NoError(t, e)

	e = target.AddCHAP(&CHAP{
		CHAPUser:     "abca",
		CHAPPassword: "abca",
	})
	assert.Error(t, e)

}

func TestTargetCfg_Delete(t *testing.T) {
	target := TargetCfg{
		TargetToolCli: "tgt-admin",
		TargetId:      "99",
		TargetIQN:     "iqn.test.com.local:local",
	}

	e := target.Delete()

	assert.NoError(t, e)
}
