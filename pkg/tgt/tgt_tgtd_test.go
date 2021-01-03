package tgt

import (
	"flag"
	"fmt"
	"github.com/ogre0403/iscsi-target-api/pkg/cfg"
	"github.com/stretchr/testify/assert"
	"os"
	"os/exec"
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
	flag.Parse()

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
		fmt.Println(fmt.Sprintf("cleanup: Remove volume %s", fullImgPath))
	})

}

func TestTgtd_CreateTarget(t *testing.T) {
	tc := &cfg.TargetCfg{
		TargetId:  "1",
		TargetIQN: "iqn.2017-07.com.hiroom2:aaadd",
	}

	err := mgr.CreateTarget(tc)
	assert.NoError(t, err)

	t.Cleanup(func() {
		cmd := exec.Command("/bin/sh", "-c",
			fmt.Sprintf("tgt-admin --delete tid=%s", tc.TargetId),
		)
		cmd.Run()
		fmt.Println(fmt.Sprintf("cleanup: Remove target with tid=%s", tc.TargetId))
	})

}

func Test_ValidateIQN(t *testing.T) {
	correct_iqn := "iqn.1993-08.org.debian:01:30d46bc47a7"
	b := validateIQN(correct_iqn)
	assert.Equal(t, true, b)

	wrong_iqn := "iq.2017-07.com.hiroom2:aaadd"
	b = validateIQN(wrong_iqn)
	assert.Equal(t, false, b)
}

func Test_FindMaxT(t *testing.T) {
	s := "Target 3: iqn.2017-07.com.hiroom2:aaadd\nTarget 7: iqn.2017-07.com.hiroom2:aaadd\nTarget 1: iqn.2017-07.com.hiroom2:aaadd"
	r := _findMax(s)
	assert.Equal(t, "7", r)
}
