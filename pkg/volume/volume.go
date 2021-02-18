package volume

import (
	"errors"
	"fmt"
	"github.com/ogre0403/go-lvm"
)

const (
	VolumeTypeLVM    = "lvm"
	VolumeTypeTGTIMG = "tgtimg"
)

type Volume interface {
	Create() error
	Delete() error
	Path() (string, error)
	IsExist() (bool, error)
}

type BasicVolume struct {
	Group         string `json:"group"`
	Size          uint64 `json:"size"`
	Unit          string `json:"unit"`
	Name          string `json:"name"`
	Type          string `json:"type"`
	ThinProvision bool   `json:"thin"`
}

func (v *BasicVolume) deletePreCheck() error {
	if v.Group == "" {
		return errors.New("volume group name is not defined")
	}

	if v.Name == "" {
		return errors.New("logical volume name is not defined")
	}

	return nil
}

func (v *BasicVolume) provisionPreCheck() error {

	if err := v.deletePreCheck(); err != nil {
		return err
	}

	if v.Size == 0 {
		return errors.New("logical volume size is not defined")
	}

	if _, exist := lvm.SizeUnit[v.Unit]; !exist {
		return errors.New(fmt.Sprintf("%s is not supported unit", v.Unit))
	}

	return nil

}
