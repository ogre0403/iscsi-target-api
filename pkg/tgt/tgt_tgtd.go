package tgt

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	log "github.com/golang/glog"
	"github.com/ogre0403/iscsi-target-api/pkg/cfg"
	"os"
	"os/exec"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

const (
	TGTADM = "tgtadm"
	TGTIMG = "tgtimg"
)

type tgtd struct {
	BaseImagePath string
	tgtimgCmd     string
	tgtadmCmd     string
}

func newTgtdTarget(mgrCfg *cfg.ManagerCfg) (TargetManager, error) {

	exist, tgtadm, tgtimg, e := isCmdExist()
	if exist == false {
		return nil, e
	}

	log.Info(fmt.Sprintf("found %s in %s", TGTADM, tgtadm))
	log.Info(fmt.Sprintf("found %s in %s", TGTIMG, tgtimg))

	return &tgtd{
		BaseImagePath: mgrCfg.BaseImagePath,
		tgtadmCmd:     tgtadm,
		tgtimgCmd:     tgtimg,
	}, nil
}

func isCmdExist() (bool, string, string, error) {

	var stdout bytes.Buffer
	cmd := exec.Command("/bin/sh", "-c", "command -v "+TGTADM)
	cmd.Stdout = &stdout

	if err := cmd.Run(); err != nil {
		return false, "", "", errors.New(fmt.Sprintf("%s not found", TGTADM))
	}
	tgtadmCmd := strings.TrimSpace(string(stdout.Bytes()))

	var stdout1 bytes.Buffer
	cmd = exec.Command("/bin/sh", "-c", "command -v "+TGTIMG)
	cmd.Stdout = &stdout1
	if err := cmd.Run(); err != nil {
		return false, "", "", errors.New(fmt.Sprintf("%s not found", TGTIMG))
	}
	tgtimgCmd := strings.TrimSpace(string(stdout1.Bytes()))

	return true, tgtadmCmd, tgtimgCmd, nil
}

// todo: support LVM based volume
func (t *tgtd) CreateVolume(cfg *cfg.VolumeCfg) error {

	fullImgPath := t.BaseImagePath + "/" + cfg.Path + "/" + cfg.Name

	if _, err := os.Stat(fullImgPath); !os.IsNotExist(err) {
		return errors.New(fmt.Sprintf("Image %s alreay exist", fullImgPath))
	}

	var stdout, stderr bytes.Buffer

	cmd := exec.Command("/bin/sh", "-c",
		fmt.Sprintf("%s --op new --device-type disk --type disk --size %s --file %s ", t.tgtimgCmd, cfg.Size, fullImgPath),
	)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return errors.New(fmt.Sprintf(string(stderr.Bytes())))
	}
	log.Info(string(stdout.Bytes()))

	return nil
}

func (t *tgtd) CreateTarget(cfg *cfg.TargetCfg) error {

	if !validateIQN(cfg.TargetIQN) {
		return errors.New(fmt.Sprintf("%s is not valid IQN format", cfg.TargetIQN))
	}

	tid := cfg.TargetId
	if cfg.TargetId == "" {
		tid = maxTargetId()
		if tid == "-1" {
			return errors.New("Can not find maximun target id")
		}
	}

	var stdout, stderr bytes.Buffer

	cmd := exec.Command("/bin/sh", "-c",
		fmt.Sprintf("%s --lld iscsi --op new --mode target --tid %s -T %s ", t.tgtadmCmd, tid, cfg.TargetIQN),
	)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return errors.New(fmt.Sprintf(string(stderr.Bytes())))
	}

	log.Info(fmt.Sprintf("Created Target %s with tid %s", cfg.TargetIQN, tid))
	return nil
}
func (t *tgtd) CreateLun(cfg *cfg.LunCfg) error {
	return nil
}
func (t *tgtd) AttachLun() error {
	return nil
}
func (t *tgtd) Reload() error {
	return nil
}

func validateIQN(iqn string) bool {
	r, _ := regexp.Compile("iqn\\.(\\d{4}-\\d{2})\\.([^:]+)(:)([^,:\\s']+)")
	return r.MatchString(iqn)
}

func maxTargetId() string {
	var stdout, stderr bytes.Buffer

	cmd := exec.Command("/bin/sh", "-c",
		fmt.Sprintf("tgtadm --lld iscsi --op show --mode target | grep -E \"iqn\\.[0-9]{4}-[0-9]{2}\" | grep Target "),
	)
	cmd.Stdout = &stdout

	if err := cmd.Run(); err != nil {
		log.Info(fmt.Sprintf(string(stderr.Bytes())))
		return "-1"
	}

	return _findMax(string(stdout.Bytes()))
}

func _findMax(s string) string {

	scanner := bufio.NewScanner(bufio.NewReader(strings.NewReader(s)))

	aa := []int{}
	for scanner.Scan() {
		line := scanner.Text()
		a, _ := strconv.Atoi(strings.Split(strings.Split(line, ":")[0], " ")[1])
		aa = append(aa, a)
	}

	sort.Ints(aa)

	return strconv.Itoa(aa[len(aa)-1])
}
