package tgt

import (
	"github.com/gin-gonic/gin"
	"github.com/ogre0403/iscsi-target-api/pkg/cfg"
	"github.com/ogre0403/iscsi-target-api/pkg/volume"
)

type mockupTarget struct {
	targetConf    string
	BaseImagePath string
}

func newMockupTarget(mgrCfg *cfg.ManagerCfg) (TargetManager, error) {

	t := &mockupTarget{
		BaseImagePath: mgrCfg.BaseImagePath,
		targetConf:    mgrCfg.TargetConf,
	}

	return t, nil
}

func (t *mockupTarget) CreateVolume(v *volume.BasicVolume) error {
	return nil
}

func (t *mockupTarget) AttachLun(l *cfg.LunCfg) error {
	return nil
}
func (t *mockupTarget) DeleteTarget(target *cfg.TargetCfg) error {
	return nil
}

func (t *mockupTarget) DeleteVolume(v *volume.BasicVolume) error {
	return nil
}

func (t *mockupTarget) CreateVolumeAPI(c *gin.Context) {
	responseWithOk(c)
}

func (t *mockupTarget) AttachLunAPI(c *gin.Context) {
	responseWithOk(c)
}

func (t *mockupTarget) DeleteVolumeAPI(c *gin.Context) {
	responseWithOk(c)
}

func (t *mockupTarget) DeleteTargetAPI(c *gin.Context) {
	responseWithOk(c)
}
func (t *mockupTarget) Save() error {
	return nil
}
