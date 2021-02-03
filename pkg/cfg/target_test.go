package cfg

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

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

	TgtimgVol.SetBaseImgPath(ImgPath)
	TgtimgVol.SetTgtimgCmd(TgtimgCmd)
	e := TgtimgVol.Create()
	assert.NoError(t, e)

	path, _ := TgtimgVol.Path()

	target := TargetCfg{
		TargetToolCli: "tgtadm",
		TargetId:      "99",
		TargetIQN:     "iqn.test.com.local:local",
	}

	e = target.AddLun(path)
	assert.NoError(t, e)

	TgtimgVol.Delete()

}

func TestTargetCfg_ACL(t *testing.T) {
	target := TargetCfg{
		TargetToolCli: "tgtadm",
		TargetId:      "99",
		TargetIQN:     "iqn.test.com.local:local",
	}

	e := target.SetACL([]string{})
	assert.NoError(t, e)
	e = target.SetACL([]string{"192.168.1.1","192.168.1.2"})
	assert.NoError(t, e)

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
