package tgt

import (
	log "github.com/golang/glog"
	"github.com/ogre0403/iscsi-target-api/pkg/cfg"
	"github.com/ogre0403/iscsi-target-api/pkg/volume"
)

const (
	TgtdType = "tgtd"
	// todo: support iscsitarget
	IscsiTargetType = "iscsitarget"
)

type TargetManager interface {
	CreateVolume(*volume.BasicVolume) error
	AttachLun(*cfg.LunCfg) error
	DeleteTarget(*cfg.TargetCfg) error
	DeleteVolume(*volume.BasicVolume) error
	Save() error
}

func NewTarget(targetType string, mgrCfg *cfg.ManagerCfg) (TargetManager, error) {
	switch targetType {
	case TgtdType:
		log.Infof("Initialize %s target manager tool", TgtdType)
		return newTgtdTarget(mgrCfg)
	default:
		log.Infof("%s is not supported tool, use %s as default manager tool", targetType, TgtdType)
		return newTgtdTarget(mgrCfg)
	}
}
