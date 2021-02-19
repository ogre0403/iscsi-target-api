package manager

import (
	log "github.com/golang/glog"
	"github.com/ogre0403/iscsi-target-api/pkg/model"
	"github.com/ogre0403/iscsi-target-api/pkg/target"
	"github.com/ogre0403/iscsi-target-api/pkg/volume"
)

const (
	TgtdType = "tgtd"
	// todo: support iscsitarget
	IscsiTargetType = "iscsitarget"
)

type TargetManager interface {
	CreateVolume(*volume.BasicVolume) error
	AttachLun(*model.Lun) error
	DeleteTarget(*target.Target) error
	DeleteVolume(*volume.BasicVolume) error
	Save() error
}

func NewTarget(targetType string, mgrCfg *model.ManagerCfg) (TargetManager, error) {
	switch targetType {
	case TgtdType:
		log.Infof("Initialize %s target manager tool", TgtdType)
		return newTgtdTarget(mgrCfg)
	default:
		log.Infof("%s is not supported tool, use %s as default manager tool", targetType, TgtdType)
		return newTgtdTarget(mgrCfg)
	}
}
