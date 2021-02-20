package target

import (
	"bytes"
	"errors"
	"fmt"
	log "github.com/golang/glog"
	"github.com/ogre0403/iscsi-target-api/pkg/model"
	"os/exec"
	"strconv"
	"strings"
)

const TgtadmCmd = "tgtadm"

type TgtdTarget struct {
	BasicTarget
}

func NewTgtdTarget(basic *BasicTarget) *TgtdTarget {

	result := &TgtdTarget{
		BasicTarget: BasicTarget{
			TargetIQN: basic.TargetIQN,
		},
	}
	// try to find tiq for iqn
	tid := queryTargetId(basic.TargetIQN)

	// if tid is not found, allocate next tid for iqn
	if tid == "-1" {
		nextTid := queryMaxTargetId() + 1
		result.TargetId = strconv.Itoa(nextTid)
		return result
	}

	result.TargetId = tid
	return result
}

// create target
// eg. tgtadm --lld iscsi --op new --mode target --tid 1 -T iqn.2017-07.com.hiroom2:debian-9
func (t *TgtdTarget) Create() error {
	var stdout, stderr bytes.Buffer

	cmd := exec.Command("/bin/sh", "-c",
		fmt.Sprintf("%s --lld iscsi --op new --mode target --tid %s -T %s ", TgtadmCmd, t.TargetId, t.TargetIQN),
	)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	log.V(3).Info(cmd.String())
	if err := cmd.Run(); err != nil {
		return errors.New(strings.TrimSpace(fmt.Sprintf(string(stderr.Bytes()))))
	}
	log.Info(string(stdout.Bytes()))
	return nil
}

// delete target
// eg. tgtadm --lld iscsi --op delete --mode target  --tid 1
func (t *TgtdTarget) Delete() error {

	tid := queryTargetId(t.TargetIQN)
	if tid == "-1" {
		return errors.New(fmt.Sprintf("target %s not found", t.TargetIQN))
	}

	var stdout, stderr bytes.Buffer
	cmd := exec.Command("/bin/sh", "-c",
		fmt.Sprintf("%s --lld iscsi --op delete --mode target --tid %s  ", TgtadmCmd, t.TargetId),
	)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	log.V(3).Info(cmd.String())
	if err := cmd.Run(); err != nil {
		return errors.New(strings.TrimSpace(fmt.Sprintf(string(stderr.Bytes()))))
	}

	log.Info(string(stdout.Bytes()))
	return nil
}

// setup LUN
// eg. tgtadm --lld iscsi --op new --mode logicalunit --tid 1 --lun 1 -b /var/lib/iscsi/10m-$i.img
func (t *TgtdTarget) AddLun(volPath string) error {
	var stdout, stderr bytes.Buffer

	cmd := exec.Command("/bin/sh", "-c",
		fmt.Sprintf("%s --lld iscsi --op new --mode logicalunit --tid %s --lun 1 -b %s ", TgtadmCmd, t.TargetId, volPath),
	)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	log.V(3).Info(cmd.String())
	if err := cmd.Run(); err != nil {
		return errors.New(strings.TrimSpace(fmt.Sprintf(string(stderr.Bytes()))))
	}
	log.Info(string(stdout.Bytes()))

	return nil
}

// setup ACL
// eg. tgtadm --lld iscsi --op bind --mode target --tid 1 -I 192.168.1.1
func (t *TgtdTarget) AddACL(aclList []string) error {
	var stdout, stderr bytes.Buffer

	if len(aclList) == 0 {
		cmd := exec.Command("/bin/sh", "-c",
			fmt.Sprintf("%s --lld iscsi --op bind --mode target --tid %s -I %s ", TgtadmCmd, t.TargetId, "ALL"),
		)
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr
		log.V(3).Info(cmd.String())
		if err := cmd.Run(); err != nil {
			return errors.New(strings.TrimSpace(fmt.Sprintf(string(stderr.Bytes()))))
		}
		log.Info(string(stdout.Bytes()))
		return nil
	}

	for _, ip := range aclList {
		cmd := exec.Command("/bin/sh", "-c",
			fmt.Sprintf("%s --lld iscsi --op bind --mode target --tid %s -I %s ", TgtadmCmd, t.TargetId, ip),
		)
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr
		log.V(3).Info(cmd.String())
		if err := cmd.Run(); err != nil {
			return errors.New(strings.TrimSpace(fmt.Sprintf(string(stderr.Bytes()))))
		}
		log.Info(string(stdout.Bytes()))
	}
	return nil
}

// set CHAP
// eg. tgtadm --lld iscsi --op bind --mode account --tid 1 --user benjr --outgoing
//     tgtadm --lld iscsi --op bind --mode account --tid 1 --user benjr
func (t *TgtdTarget) AddCHAP(chap *model.CHAP) error {
	var stdout, stderr bytes.Buffer

	if chap.CHAPUser != "" && chap.CHAPPassword != "" {
		cmd := exec.Command("/bin/sh", "-c",
			fmt.Sprintf("%s --lld iscsi --op bind --mode account --tid %s --user %s ",
				TgtadmCmd, t.TargetId, chap.CHAPUser),
		)
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr
		log.V(3).Info(cmd.String())
		if err := cmd.Run(); err != nil {
			return errors.New(strings.TrimSpace(fmt.Sprintf(string(stderr.Bytes()))))
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
				TgtadmCmd, t.TargetId, chap.CHAPUserIn),
		)
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr
		log.V(3).Info(cmd.String())
		if err := cmd.Run(); err != nil {
			return errors.New(strings.TrimSpace(fmt.Sprintf(string(stderr.Bytes()))))
		}
		log.Info(string(stdout.Bytes()))
	} else if chap.CHAPUserIn == "" && chap.CHAPPasswordIn == "" {
		log.Warningf("both chap username_in and chap password_in are empty, imply disable chap")
	} else {
		return errors.New("CHAP username_in or password_in is empty")
	}

	return nil
}
