package tgt

import (
	"github.com/ogre0403/iscsi-target-api/pkg/cfg"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

var mgr tgtd

func init() {

	mgr = tgtd{
		BaseImagePath: "/var/lib/iscsi/",
		tgtadmCmd:     "tgtadm",
		tgtimgCmd:     "tgtimg",
	}
}

func TestNewTarget(t *testing.T) {

	_, err := newTgtdTarget(&cfg.ManagerCfg{
		BaseImagePath: "/var/lib/iscsi/",
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
