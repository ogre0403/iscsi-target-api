package client

import (
	"github.com/ogre0403/iscsi-target-api/pkg/cfg"
	"github.com/ogre0403/iscsi-target-api/pkg/rest"
	"github.com/ogre0403/iscsi-target-api/pkg/tgt"
	"github.com/stretchr/testify/assert"
	"testing"
)

var s *rest.APIServer
var c *Client

func init() {

	m, _ := tgt.NewTarget(tgt.MockupType, &cfg.ManagerCfg{})
	s = rest.NewAPIServer(m)
	c = NewClient("127.0.0.1", 80)
	go s.RunServer(80)

}

func Test_ClientCreateVolume(t *testing.T) {

	err := c.CreateVolume(&cfg.VolumeCfg{})
	assert.NoError(t, err)
}

func Test_ClientDeleteVolume(t *testing.T) {

	err := c.DeleteVolume(&cfg.VolumeCfg{})
	assert.NoError(t, err)
}

func Test_ClientAttachLun(t *testing.T) {
	err := c.AttachLun(&cfg.LunCfg{})
	assert.NoError(t, err)
}

func Test_ClientDeleteTarget(t *testing.T) {

	err := c.DeleteTarget(&cfg.TargetCfg{})
	assert.NoError(t, err)
}
