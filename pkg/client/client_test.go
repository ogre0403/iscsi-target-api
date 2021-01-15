package client

import (
	"errors"
	"github.com/ogre0403/iscsi-target-api/pkg/cfg"
	"github.com/ogre0403/iscsi-target-api/pkg/rest"
	"github.com/ogre0403/iscsi-target-api/pkg/tgt"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

var s *rest.APIServer
var c, wrong_c *Client

func init() {

	m, _ := tgt.NewTarget(tgt.MockupType, &cfg.ManagerCfg{})
	sc := &cfg.ServerCfg{
		Port:     80,
		Username: "admin",
		Password: "password",
	}

	wrong_sc := &cfg.ServerCfg{
		Port:     80,
		Username: "admin",
		Password: "passwor",
	}

	s = rest.NewAPIServer(m, sc)
	c = NewClient("127.0.0.1", sc)

	wrong_c = NewClient("127.0.0.1", wrong_sc)

	go s.RunServer()
	time.Sleep(3 * time.Second)
}

func Test_ClientCreateVolume(t *testing.T) {

	err := c.CreateVolume(&cfg.VolumeCfg{})
	assert.NoError(t, err)

	err = wrong_c.CreateVolume(&cfg.VolumeCfg{})
	if assert.Error(t, err) {
		assert.Equal(t, errors.New("request is not authorized"), err)
	}

}

func Test_ClientDeleteVolume(t *testing.T) {

	err := c.DeleteVolume(&cfg.VolumeCfg{})
	assert.NoError(t, err)

	err = wrong_c.DeleteVolume(&cfg.VolumeCfg{})
	if assert.Error(t, err) {
		assert.Equal(t, errors.New("request is not authorized"), err)
	}
}

func Test_ClientAttachLun(t *testing.T) {
	err := c.AttachLun(&cfg.LunCfg{})
	assert.NoError(t, err)

	err = wrong_c.AttachLun(&cfg.LunCfg{})
	if assert.Error(t, err) {
		assert.Equal(t, errors.New("request is not authorized"), err)
	}
}

func Test_ClientDeleteTarget(t *testing.T) {

	err := c.DeleteTarget(&cfg.TargetCfg{})
	assert.NoError(t, err)

	err = wrong_c.DeleteTarget(&cfg.TargetCfg{})
	if assert.Error(t, err) {
		assert.Equal(t, errors.New("request is not authorized"), err)
	}
}
