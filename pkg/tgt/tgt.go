package tgt

import (
	log "github.com/golang/glog"
	"github.com/ogre0403/iscsi-target-api/pkg/cfg"
)

const (
	tgtdType = "tgtd"
	// todo: support iscsitarget
	iScsiTargetType = "iscsitarget"
)

type TargetManager interface {
	CreateVolume(cfg *cfg.VolumeCfg) error
	CreateTarget(cfg *cfg.TargetCfg) error
	AttachLun(cfg *cfg.LunCfg) error
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
