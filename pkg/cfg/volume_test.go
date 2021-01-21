package cfg

import (
	"errors"
	"fmt"
	"github.com/ogre0403/go-lvm"
	"github.com/stretchr/testify/assert"
	"testing"
)

var wrongVol = VolumeCfg{
	Type:  "test",
	Group: "test",
	Name:  "test",
}

var LVMVol = VolumeCfg{
	Type:  VolumeTypeLVM,
	Group: "vg-0",
	Name:  "test",
	Size:  12,
	Unit:  lvm.MiB,
}

var TgtimgVol = VolumeCfg{
	Type:  VolumeTypeTGTIMG,
	Group: "test",
	Name:  "test",
	Size:  12,
	Unit:  lvm.MiB,
}

const ImgPath = "/var/lib/iscsi"
const TgtimgCmd = "tgtimg"

func TestVolumeCfg_SetBaseImgPath(t *testing.T) {
	TgtimgVol.SetBaseImgPath(ImgPath)
	TgtimgVol.SetTgtimgCmd(TgtimgCmd)
	assert.Equal(t, TgtimgVol.baseImagePath, ImgPath)
	assert.Equal(t, TgtimgVol.tgtimgCmd, TgtimgCmd)
}

func TestVolume_Path(t *testing.T) {

	p, err := wrongVol.Path()
	assert.Equal(t, "", p)
	assert.Error(t, errors.New(fmt.Sprintf("%s is not supported volume provision tool", wrongVol.Type)), err)

	p, err = LVMVol.Path()
	assert.Equal(t, "/dev/vg-0/test", p)
	assert.NoError(t, err)
	e, err := LVMVol.IsExist()
	assert.Equal(t, false, e)

	p, err = TgtimgVol.Path()
	assert.Equal(t, "/var/lib/iscsi/test/test", p)
	assert.NoError(t, err)
	e, err = TgtimgVol.IsExist()
	assert.Equal(t, false, e)
}

func TestVolumeCfg_tgtimgProvision(t *testing.T) {
	e := TgtimgVol.Create()
	assert.NoError(t, e)

	t.Cleanup(func() {
		err := TgtimgVol.tgtimgDelete()
		assert.NoError(t, err)
	})

}

func TestVolumeCfg_lvmProvision(t *testing.T) {
	e := LVMVol.Create()
	assert.NoError(t, e)

	t.Cleanup(func() {
		err := LVMVol.lvmDelete()
		assert.NoError(t, err)
	})
}

func TestVolumeCfg_lvmThinProvision(t *testing.T) {
	LVMVol.ThinProvision = true
	e := LVMVol.Create()
	assert.Error(t, e)
}
