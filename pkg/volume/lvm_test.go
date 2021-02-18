package volume

import (
	"github.com/ogre0403/go-lvm"
	"github.com/stretchr/testify/assert"
	"testing"
)

var LVMVol = LvmVolume{
	BasicVolume: BasicVolume{
		Type:  VolumeTypeLVM,
		Group: "vg-0",
		Name:  "test",
		Size:  12,
		Unit:  lvm.MiB,
	},
}

func init() {
	lvm.Initialize()
}

func TestLvmVolume_Path(t *testing.T) {
	p, err := LVMVol.Path()
	assert.Equal(t, "/dev/vg-0/test", p)
	assert.NoError(t, err)
	e, err := LVMVol.IsExist()
	assert.Equal(t, false, e)
}

func TestLvmVolume_Create(t *testing.T) {
	e := LVMVol.Create()
	assert.NoError(t, e)

	t.Cleanup(func() {
		err := LVMVol.Delete()
		assert.NoError(t, err)
	})
}

func TestLvmVolume_CreateThin(t *testing.T) {
	LVMVol.ThinProvision = true
	e := LVMVol.Create()
	assert.Error(t, e)

	LVMVol.ThinPool = "pool0"
	e = LVMVol.Create()
	assert.NoError(t, e)

	vvv, e := lvm.BD_LVM_LvInfo(LVMVol.Group, LVMVol.Name)
	assert.NoError(t, e)
	assert.Equal(t, vvv.IsThinVolume(), true)

	t.Cleanup(func() {
		err := LVMVol.Delete()
		assert.NoError(t, err)
	})

}
