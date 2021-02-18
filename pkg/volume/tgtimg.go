package volume

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

type ImageVolume struct {
	BasicVolume
	TgtimgCmd     string
	BaseImagePath string
}

func (v *ImageVolume) Create() error {

	if err := v.provisionPreCheck(); err != nil {
		return err
	}

	if _, err := v.Path(); err != nil {
		return err
	}

	sizeUnit := fmt.Sprintf("%dm", uint64(lvm.UnitTranslate(v.Size, v.Unit, lvm.MiB)))

	// already check in preCheck
	fullImgPath, _ := v.Path()

	if exist, _ := v.IsExist(); exist {
		return errors.New(fmt.Sprintf("Image %s alreay exist", fullImgPath))
	}

	if _, err := os.Stat(v.BaseImagePath + "/" + v.Group); os.IsNotExist(err) {
		if err := os.MkdirAll(v.BaseImagePath+"/"+v.Group, 0755); err != nil {
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
		fmt.Sprintf("%s --op new --device-type disk --type disk --size %s %s --file %s ", v.TgtimgCmd, sizeUnit, thin, fullImgPath),
	)

	log.V(3).Info(cmd.String())
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return errors.New(fmt.Sprintf(string(stderr.Bytes())))
	}
	log.V(2).Info(string(stdout.Bytes()))
	return nil
}

func (v *ImageVolume) Delete() error {
	if err := v.deletePreCheck(); err != nil {
		return err
	}

	volSubDir := v.BaseImagePath + "/" + v.Group
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

func (v *ImageVolume) Path() (string, error) {
	return v.BaseImagePath + "/" + v.Group + "/" + v.Name, nil
}

func (v *ImageVolume) IsExist() (bool, error) {
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
