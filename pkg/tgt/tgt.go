package tgt

import (
	"github.com/gin-gonic/gin"
	log "github.com/golang/glog"
	"github.com/ogre0403/iscsi-target-api/pkg/cfg"
)

const (
	tgtdType = "tgtd"
	// todo: support iscsitarget
	iScsiTargetType = "iscsitarget"
)

type TargetManager interface {
	CreateVolume(*cfg.VolumeCfg) error
	CreateVolumeAPI(*gin.Context)
	AttachLun(*cfg.LunCfg) error
	AttachLunAPI(*gin.Context)
	DeleteTarget(*cfg.TargetCfg) error
	DeleteTargetAPI(*gin.Context)
	DeleteVolume(*cfg.VolumeCfg) error
	DeleteVolumeAPI(*gin.Context)
	Reload() error
}

func NewTarget(targetType string, mgrCfg *cfg.ManagerCfg) (TargetManager, error) {
	switch targetType {
	case tgtdType:
		log.Infof("Initialize %s target manager tool", tgtdType)
		return newTgtdTarget(mgrCfg)
	default:
		log.Infof("%s is not supported tool, use %s as default manager tool", targetType, tgtdType)
		return newTgtdTarget(mgrCfg)
	}
}
