package target

import "github.com/ogre0403/iscsi-target-api/pkg/model"

type Target interface {
	Create() error
	Delete() error
	AddLun(volPath string) error
	AddACL(aclList []string) error
	AddCHAP(chap *model.CHAP) error
}

type BasicTarget struct {
	TargetId  string `json:"-"`
	TargetIQN string `json:"targetIQN"`
}
