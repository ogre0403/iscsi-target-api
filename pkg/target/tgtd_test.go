package target

import (
	"github.com/ogre0403/go-lvm"
	"github.com/ogre0403/iscsi-target-api/pkg/model"
	"github.com/ogre0403/iscsi-target-api/pkg/volume"
	"github.com/stretchr/testify/assert"
	"testing"
)

var tgtimgVol = volume.ImageVolume{
	BasicVolume: volume.BasicVolume{
		Type:  volume.VolumeTypeTGTIMG,
		Group: "test1",
		Name:  "test1",
		Size:  12,
		Unit:  lvm.MiB,
	},

	TgtimgCmd:     "tgtimg",
	BaseImagePath: "/var/lib/iscsi",
}

func TestTargetCfg_Create(t *testing.T) {

	target := TgtdTarget{
		BasicTarget: BasicTarget{
			TargetId:  "99",
			TargetIQN: "iqn.test.com.local:local",
		},
	}

	e := target.Create()

	assert.NoError(t, e)

}

func TestTargetCfg_AddLun(t *testing.T) {

	e := tgtimgVol.Create()
	assert.NoError(t, e)

	path, _ := tgtimgVol.Path()

	target := TgtdTarget{
		BasicTarget: BasicTarget{
			TargetId:  "99",
			TargetIQN: "iqn.test.com.local:local",
		},
	}

	e = target.AddLun(path)
	assert.NoError(t, e)

	e = tgtimgVol.Delete()
	assert.NoError(t, e)
}

func TestTargetCfg_ACL(t *testing.T) {
	target := TgtdTarget{
		BasicTarget: BasicTarget{
			TargetId:  "99",
			TargetIQN: "iqn.test.com.local:local",
		},
	}

	e := target.AddACL([]string{})
	assert.NoError(t, e)
	e = target.AddACL([]string{"192.168.1.1", "192.168.1.2"})
	assert.NoError(t, e)

}

func TestTargetCfg_AddCHAP(t *testing.T) {
	target := TgtdTarget{
		BasicTarget: BasicTarget{
			TargetId:  "99",
			TargetIQN: "iqn.test.com.local:local",
		},
	}

	e := target.AddCHAP(&model.CHAP{})
	assert.NoError(t, e)

	e = target.AddCHAP(&model.CHAP{
		CHAPUser:     "abca",
		CHAPPassword: "abca",
	})
	assert.Error(t, e)

}

func TestTargetCfg_Delete(t *testing.T) {
	target := TgtdTarget{
		BasicTarget: BasicTarget{
			TargetId:  "99",
			TargetIQN: "iqn.test.com.local:local",
		},
	}

	e := target.Delete()

	assert.NoError(t, e)
}
