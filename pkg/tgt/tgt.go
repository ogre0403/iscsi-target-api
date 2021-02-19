package tgt

import (
	"github.com/gin-gonic/gin"
	log "github.com/golang/glog"
	"github.com/ogre0403/iscsi-target-api/pkg/cfg"
	"github.com/ogre0403/iscsi-target-api/pkg/volume"
)

const (
	TgtdType   = "tgtd"
	MockupType = "mockup"
	// todo: support iscsitarget
	IscsiTargetType = "iscsitarget"
)

type TargetManager interface {
	CreateVolume(*volume.BasicVolume) error
	CreateVolumeAPI(*gin.Context)
	AttachLun(*cfg.LunCfg) error
	AttachLunAPI(*gin.Context)
	DeleteTarget(*cfg.TargetCfg) error
	DeleteTargetAPI(*gin.Context)
	DeleteVolume(*volume.BasicVolume) error
	DeleteVolumeAPI(*gin.Context)
	Save() error
}

func NewTarget(targetType string, mgrCfg *cfg.ManagerCfg) (TargetManager, error) {
	switch targetType {
	case TgtdType:
		log.Infof("Initialize %s target manager tool", TgtdType)
		return newTgtdTarget(mgrCfg)
	case MockupType:
		log.Infof("Initialize %s target manager tool", MockupType)
		return newMockupTarget(mgrCfg)
	default:
		log.Infof("%s is not supported tool, use %s as default manager tool", targetType, TgtdType)
		return newTgtdTarget(mgrCfg)
	}
}
