package manager

import (
	"errors"
	"fmt"
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
	DeleteTarget(*target.BasicTarget) error
	DeleteVolume(*volume.BasicVolume) error
	Save() error
}

func NewManager(targetType string, mgrCfg *model.ManagerCfg) (TargetManager, error) {
	switch targetType {
	case TgtdType:
		log.Infof("Initialize %s target manager tool", TgtdType)
		return newTgtdMgr(mgrCfg)
	case IscsiTargetType:
		log.Infof("Initialize %s target manager tool", IscsiTargetType)
		return nil, errors.New(fmt.Sprintf("%s not implemeted", IscsiTargetType))
	default:
		log.Infof("%s is not supported tool, use %s as default manager tool", targetType, TgtdType)
		return newTgtdMgr(mgrCfg)
	}
}
