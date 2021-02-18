package volume

import (
	"errors"
	"fmt"
	log "github.com/golang/glog"
	"github.com/ogre0403/go-lvm"
	"os"
)

type LvmVolume struct {
	BasicVolume
	ThinPool string `json:"pool"`
}

func (v *LvmVolume) Create() error {
	if err := v.provisionPreCheck(); err != nil {
		return err
	}

	if _, err := v.Path(); err != nil {
		return err
	}

	if v.ThinProvision {
		if v.ThinPool == "" {
			return errors.New(
				fmt.Sprintf("LVM volume %s/%s use thin provision, but thin pool is not defined ",
					v.Group, v.Name))
		}
		log.V(2).Infof("thin provision LVM volume %s/%s", v.ThinPool, v.Name)
		return lvm.BD_LVM_ThLvCreate(v.Group, v.ThinPool, v.Name, lvm.HumanReadableToBytes(v.Size, v.Unit))
	} else {
		return lvm.BD_LVM_LvCreate(v.Group, v.Name, lvm.HumanReadableToBytes(v.Size, v.Unit))
	}
}

func (v *LvmVolume) Delete() error {
	if err := v.deletePreCheck(); err != nil {
		return err
	}
	return lvm.BD_LVM_LvRemove(v.Group, v.Name)
}

func (v *LvmVolume) Path() (string, error) {
	return "/dev/" + v.Group + "/" + v.Name, nil
}

func (v *LvmVolume) IsExist() (bool, error) {
	p, err := v.Path()
	if err != nil {
		return false, err
	}

	if _, err := os.Stat(p); os.IsNotExist(err) {
		return false, err
	}
	return true, nil
}
