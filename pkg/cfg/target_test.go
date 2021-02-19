package cfg

import (
	"github.com/ogre0403/go-lvm"
	"github.com/ogre0403/iscsi-target-api/pkg/volume"
	"github.com/stretchr/testify/assert"
	"testing"
)

var tgtimgVol = volume.ImageVolume{
	BasicVolume: volume.BasicVolume{
		Type:  volume.VolumeTypeTGTIMG,
		Group: "test",
		Name:  "test",
		Size:  12,
		Unit:  lvm.MiB,
	},

	TgtimgCmd:     "tgtimg",
	BaseImagePath: "/var/lib/iscsi",
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

	e := tgtimgVol.Create()
	assert.NoError(t, e)

	path, _ := tgtimgVol.Path()

	target := TargetCfg{
		TargetToolCli: "tgtadm",
		TargetId:      "99",
		TargetIQN:     "iqn.test.com.local:local",
	}

	e = target.AddLun(path)
	assert.NoError(t, e)

	tgtimgVol.Delete()
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
