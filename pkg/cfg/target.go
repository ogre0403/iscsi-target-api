package cfg

import (
	"bytes"
	"errors"
	"fmt"
	log "github.com/golang/glog"
	"os/exec"
)

type TargetCfg struct {
	TargetToolCli string `json:"-"`
	TargetId      string `json:"-"`
	TargetIQN     string `json:"targetIQN"`
}

// create target
// eg. tgtadm --lld iscsi --op new --mode target --tid 1 -T iqn.2017-07.com.hiroom2:debian-9
func (t *TargetCfg) Create() error {
	var stdout, stderr bytes.Buffer

	cmd := exec.Command("/bin/sh", "-c",
		fmt.Sprintf("%s --lld iscsi --op new --mode target --tid %s -T %s ", t.TargetToolCli, t.TargetId, t.TargetIQN),
	)
	log.Info(cmd.String())

	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return errors.New(fmt.Sprintf(string(stderr.Bytes())))
	}
	log.Info(string(stdout.Bytes()))
	return nil
}

// delete target
// eg. tgt-admin --delete tid=1
func (t *TargetCfg) Delete() error {

	var stdout, stderr bytes.Buffer
	cmd := exec.Command("/bin/sh", "-c",
		fmt.Sprintf("%s --delete tid=%s", t.TargetToolCli, t.TargetId),
	)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return errors.New(fmt.Sprintf(string(stderr.Bytes())))
	}

	log.Info(string(stdout.Bytes()))
	return nil
}

// setup LUN
// eg. tgtadm --lld iscsi --op new --mode logicalunit --tid 1 --lun 1 -b /var/lib/iscsi/10m-$i.img
func (t *TargetCfg) AddLun(volPath string) error {
	var stdout, stderr bytes.Buffer

	cmd := exec.Command("/bin/sh", "-c",
		fmt.Sprintf("%s --lld iscsi --op new --mode logicalunit --tid %s --lun 1 -b %s ", t.TargetToolCli, t.TargetId, volPath),
	)
	log.Info(cmd.String())

	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return errors.New(fmt.Sprintf(string(stderr.Bytes())))
	}
	log.Info(string(stdout.Bytes()))

	return nil
}

// setup ACL
// eg. tgtadm --lld iscsi --op bind --mode target --tid 1 -I 192.168.1.1
func (t *TargetCfg) SetACL(aclList []string) error {
	var stdout, stderr bytes.Buffer

	if len(aclList) == 0 {
		cmd := exec.Command("/bin/sh", "-c",
			fmt.Sprintf("%s --lld iscsi --op bind --mode target --tid %s -I %s ", t.TargetToolCli, t.TargetId, "ALL"),
		)
		log.Info(cmd.String())

		cmd.Stdout = &stdout
		cmd.Stderr = &stderr
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
		log.Info(cmd.String())

		cmd.Stdout = &stdout
		cmd.Stderr = &stderr
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
func (t *TargetCfg) AddCHAP(chap *CHAP) error {
	//if chap.CHAPUser == "" || chap.CHAPPassword == "" {
	//	return errors.New("")
	//}
	//
	//if chap.CHAPUserIn == "" || chap.CHAPPasswordIn == "" {
	//	return errors.New("")
	//}

	return nil
}
