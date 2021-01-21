package cfg

import (
	"bytes"
	"errors"
	"fmt"
	log "github.com/golang/glog"
	"github.com/ogre0403/go-lvm"
	"io"
	"os"
	"os/exec"
)

const (
	VolumeTypeLVM    = "lvm"
	VolumeTypeTGTIMG = "tgtimg"
)

type VolumeCfg struct {
	Group         string `json:"group"`
	Size          uint64 `json:"size"`
	Unit          string `json:"unit"`
	Name          string `json:"name"`
	Type          string `json:"type"`
	ThinProvision bool   `json:"thin"`

	// These fields are target specified, should be defined in target manager
	baseImagePath string
	tgtimgCmd     string
	thinPool      string
}

func (v *VolumeCfg) SetBaseImgPath(path string) {
	v.baseImagePath = path
}

func (v *VolumeCfg) SetTgtimgCmd(cmd string) {
	v.tgtimgCmd = cmd
}

func (v *VolumeCfg) SetThinPool(pool string) {
	v.thinPool = pool
}

func (v *VolumeCfg) Create() error {

	switch v.Type {
	case VolumeTypeTGTIMG:
		log.V(2).Infof("Provision volume by %s ", VolumeTypeTGTIMG)
		return v.tgtimgProvision()
	case VolumeTypeLVM:
		log.V(2).Infof("Provision volume by %s ", VolumeTypeLVM)
		return v.lvmProvision()
	default:
		log.Errorf("%s is not supported volume provision tool", v.Type)
		return errors.New(fmt.Sprintf("%s is not supported volume provision tool", v.Type))
	}
}

func (v *VolumeCfg) deletePrecheck() error {
	if v.Group == "" {
		return errors.New("volume group name is not defined")
	}

	if v.Name == "" {
		return errors.New("logical volume name is not defined")
	}

	return nil
}

func (v *VolumeCfg) provisionPrecheck() error {

	if err := v.deletePrecheck(); err != nil {
		return err
	}

	if v.Size == 0 {
		return errors.New("logical volume size is not defined")
	}

	if _, exist := lvm.SizeUnit[v.Unit]; !exist {
		return errors.New(fmt.Sprintf("%s is not supported unit", v.Unit))
	}

	_, err := v.Path()
	return err

}

func (v *VolumeCfg) lvmProvision() error {

	if err := v.provisionPrecheck(); err != nil {
		return err
	}

	vgo, err := lvm.VgOpen(v.Group, "w")

	if err != nil {
		return err
	}

	defer vgo.Close()

	if v.ThinProvision {
		if v.thinPool == "" {
			return errors.New(
				fmt.Sprintf("LVM volume %s/%s use thin provision, but thin pool is not defined ",
					v.Group, v.Name))
		}
		_, err = vgo.CreateLvThin(v.thinPool, v.Name, uint64(lvm.UnitTranslate(v.Size, v.Unit, lvm.B)))
		log.V(2).Infof("thin provision LVM volume %s/%s", v.thinPool, v.Name)
		return err
	} else {
		_, err = vgo.CreateLvLinear(v.Name, uint64(lvm.UnitTranslate(v.Size, v.Unit, lvm.B)))
		return err
	}
}

func (v *VolumeCfg) tgtimgProvision() error {

	if err := v.provisionPrecheck(); err != nil {
		return err
	}

	sizeUnit := fmt.Sprintf("%dm", uint64(lvm.UnitTranslate(v.Size, v.Unit, lvm.MiB)))

	// already check in preCheck
	fullImgPath, _ := v.Path()

	if exist, _ := v.IsExist(); exist {
		return errors.New(fmt.Sprintf("Image %s alreay exist", fullImgPath))
	}

	if _, err := os.Stat(v.baseImagePath + "/" + v.Group); os.IsNotExist(err) {
		if err := os.MkdirAll(v.baseImagePath+"/"+v.Group, 0755); err != nil {
			return errors.New("unable to create directory to provision new volume: " + err.Error())
		}
	}

	var stdout, stderr bytes.Buffer

	thin := ""
	if v.ThinProvision == true {
		thin = "--thin-provisioning"
		log.V(2).Infof("thin provision %s image %s", VolumeTypeTGTIMG, fullImgPath)
	}

	cmd := exec.Command("/bin/sh", "-c",
		fmt.Sprintf("%s --op new --device-type disk --type disk --size %s %s --file %s ", v.tgtimgCmd, sizeUnit, thin, fullImgPath),
	)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return errors.New(fmt.Sprintf(string(stderr.Bytes())))
	}
	log.V(2).Info(string(stdout.Bytes()))
	return nil
}

func (v *VolumeCfg) Delete() error {
	switch v.Type {
	case VolumeTypeTGTIMG:
		log.V(2).Infof("Delete volume by %s ", VolumeTypeTGTIMG)
		return v.tgtimgDelete()
	case VolumeTypeLVM:
		log.V(2).Infof("Delete volume by %s ", VolumeTypeLVM)
		return v.lvmDelete()
	default:
		log.Errorf("%s is not supported volume provision tool", v.Type)
		return errors.New(fmt.Sprintf("%s is not supported volume provision tool", v.Type))
	}
}

func (v *VolumeCfg) tgtimgDelete() error {

	if err := v.deletePrecheck(); err != nil {
		return err
	}

	volSubDir := v.baseImagePath + "/" + v.Group
	fullImgPath := volSubDir + "/" + v.Name
	if err := os.Remove(fullImgPath); err != nil {
		return err
	}

	// remove dir when there is no volume
	if isEmpty, _ := isDirEmpty(volSubDir); isEmpty {
		if err := os.Remove(volSubDir); err != nil {
			return err
		}
	}
	return nil

}

func (v *VolumeCfg) lvmDelete() error {

	if err := v.deletePrecheck(); err != nil {
		return err
	}

	vgo, err := lvm.VgOpen(v.Group, "w")
	if err != nil {
		return err
	}
	defer vgo.Close()

	lv, err := vgo.LvFromName(v.Name)
	if err != nil {
		return err
	}

	// Remove LV
	err = lv.Remove()
	if err != nil {
		return err
	}
	return nil
}

func (v *VolumeCfg) Path() (string, error) {

	switch v.Type {
	case VolumeTypeTGTIMG:
		return v.baseImagePath + "/" + v.Group + "/" + v.Name, nil
	case VolumeTypeLVM:
		return "/dev/" + v.Group + "/" + v.Name, nil
	default:
		log.Errorf("%s is not supported volume provision tool", v.Type)
		return "", errors.New(fmt.Sprintf("%s is not supported volume provision tool", v.Type))
	}

}

func (v *VolumeCfg) IsExist() (bool, error) {
	p, err := v.Path()
	if err != nil {
		return false, err
	}

	if _, err := os.Stat(p); os.IsNotExist(err) {
		return false, err
	}
	return true, nil
}

func isDirEmpty(name string) (bool, error) {
	f, err := os.Open(name)
	if err != nil {
		return false, err
	}
	defer f.Close()

	_, err = f.Readdirnames(1) // Or f.Readdir(1)
	if err == io.EOF {
		return true, nil
	}
	return false, err // Either not empty or error, suits both cases
}

// todo: support LVM resize
func (v *VolumeCfg) lvmResize() error {
	return errors.New("not implemented error")
}
