package volume

import (
	"github.com/ogre0403/go-lvm"
	"github.com/stretchr/testify/assert"
	"testing"
)

var TgtimgVol = ImageVolume{
	BasicVolume: BasicVolume{
		Type:  VolumeTypeTGTIMG,
		Group: "test",
		Name:  "test",
		Size:  12,
		Unit:  lvm.MiB,
	},

	TgtimgCmd:     "tgtimg",
	BaseImagePath: "/var/lib/iscsi",
}

func init() {
	lvm.Initialize()
}

func TestImageVolume_Path(t *testing.T) {
	p, err := TgtimgVol.Path()
	assert.Equal(t, "/var/lib/iscsi/test/test", p)
	assert.NoError(t, err)
	e, err := TgtimgVol.IsExist()
	assert.Equal(t, false, e)
}

func TestImageVolume_Create(t *testing.T) {
	e := TgtimgVol.Create()
	assert.NoError(t, e)

	t.Cleanup(func() {
		err := TgtimgVol.Delete()
		assert.NoError(t, err)
	})
}
