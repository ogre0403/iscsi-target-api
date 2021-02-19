package target

import (
	"bytes"
	"errors"
	"fmt"
	log "github.com/golang/glog"
	"github.com/ogre0403/iscsi-target-api/pkg/model"
	"os/exec"
)

type Target struct {
	TargetToolCli string `json:"-"`
	TargetId      string `json:"-"`
	TargetIQN     string `json:"targetIQN"`
}

// create target
// eg. tgtadm --lld iscsi --op new --mode target --tid 1 -T iqn.2017-07.com.hiroom2:debian-9
func (t *Target) Create() error {
	var stdout, stderr bytes.Buffer

	cmd := exec.Command("/bin/sh", "-c",
		fmt.Sprintf("%s --lld iscsi --op new --mode target --tid %s -T %s ", t.TargetToolCli, t.TargetId, t.TargetIQN),
	)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	log.V(3).Info(cmd.String())
	if err := cmd.Run(); err != nil {
		return errors.New(fmt.Sprintf(string(stderr.Bytes())))
	}
	log.Info(string(stdout.Bytes()))
	return nil
}

// delete target
// eg. tgt-admin --delete tid=1
func (t *Target) Delete() error {

	var stdout, stderr bytes.Buffer
	cmd := exec.Command("/bin/sh", "-c",
		fmt.Sprintf("%s --delete tid=%s", t.TargetToolCli, t.TargetId),
	)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	log.V(3).Info(cmd.String())
	if err := cmd.Run(); err != nil {
		return errors.New(fmt.Sprintf(string(stderr.Bytes())))
	}

	log.Info(string(stdout.Bytes()))
	return nil
}

// setup LUN
// eg. tgtadm --lld iscsi --op new --mode logicalunit --tid 1 --lun 1 -b /var/lib/iscsi/10m-$i.img
func (t *Target) AddLun(volPath string) error {
	var stdout, stderr bytes.Buffer

	cmd := exec.Command("/bin/sh", "-c",
		fmt.Sprintf("%s --lld iscsi --op new --mode logicalunit --tid %s --lun 1 -b %s ", t.TargetToolCli, t.TargetId, volPath),
	)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	log.V(3).Info(cmd.String())
	if err := cmd.Run(); err != nil {
		return errors.New(fmt.Sprintf(string(stderr.Bytes())))
	}
	log.Info(string(stdout.Bytes()))

	return nil
}

// setup ACL
// eg. tgtadm --lld iscsi --op bind --mode target --tid 1 -I 192.168.1.1
func (t *Target) SetACL(aclList []string) error {
	var stdout, stderr bytes.Buffer

	if len(aclList) == 0 {
		cmd := exec.Command("/bin/sh", "-c",
			fmt.Sprintf("%s --lld iscsi --op bind --mode target --tid %s -I %s ", t.TargetToolCli, t.TargetId, "ALL"),
		)
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr
		log.V(3).Info(cmd.String())
		if err := cmd.Run(); err != nil {
			return errors.New(fmt.Sprintf(string(stderr.Bytes())))
		}
		log.Info(string(stdout.Bytes()))
		return nil
	}

	for _, ip := range aclList {
		cmd := exec.Command("/bin/sh", "-c",
			fmt.Sprintf("%s --lld iscsi --op bind --mode target --tid %s -I %s ", t.TargetToolCli, t.TargetId, ip),
		)
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr
		log.V(3).Info(cmd.String())
		if err := cmd.Run(); err != nil {
			return errors.New(fmt.Sprintf(string(stderr.Bytes())))
		}
		log.Info(string(stdout.Bytes()))
	}
	return nil
}

// set CHAP
// eg. tgtadm --lld iscsi --op bind --mode account --tid 1 --user benjr --outgoing
//     tgtadm --lld iscsi --op bind --mode account --tid 1 --user benjr
func (t *Target) AddCHAP(chap *model.CHAP) error {
	var stdout, stderr bytes.Buffer

	if chap.CHAPUser != "" && chap.CHAPPassword != "" {
		cmd := exec.Command("/bin/sh", "-c",
			fmt.Sprintf("%s --lld iscsi --op bind --mode account --tid %s --user %s ",
				t.TargetToolCli, t.TargetId, chap.CHAPUser),
		)
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr
		log.V(3).Info(cmd.String())
		if err := cmd.Run(); err != nil {
			return errors.New(fmt.Sprintf(string(stderr.Bytes())))
		}
		log.Info(string(stdout.Bytes()))
	} else if chap.CHAPUser == "" && chap.CHAPPassword == "" {
		log.Warningf("both chap username and chap password are empty, imply disable chap")
	} else {
		return errors.New("CHAP username or password is empty")
	}

	if chap.CHAPUserIn != "" && chap.CHAPPasswordIn != "" {
		cmd := exec.Command("/bin/sh", "-c",
			fmt.Sprintf("%s --lld iscsi --op bind --mode account --tid %s --user %s --outgoing",
				t.TargetToolCli, t.TargetId, chap.CHAPUserIn),
		)
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr
		log.V(3).Info(cmd.String())
		if err := cmd.Run(); err != nil {
			return errors.New(fmt.Sprintf(string(stderr.Bytes())))
		}
		log.Info(string(stdout.Bytes()))
	} else if chap.CHAPUserIn == "" && chap.CHAPPasswordIn == "" {
		log.Warningf("both chap username_in and chap password_in are empty, imply disable chap")
	} else {
		return errors.New("CHAP username_in or password_in is empty")
	}

	return nil
}
