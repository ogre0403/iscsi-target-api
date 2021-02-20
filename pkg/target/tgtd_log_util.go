package target

import (
	"bufio"
	"bytes"
	"fmt"
	log "github.com/golang/glog"
	"os/exec"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

func queryMaxTargetId() int {
	var stdout, stderr bytes.Buffer

	cmd := exec.Command("/bin/sh", "-c",
		fmt.Sprintf("tgtadm --lld iscsi --op show --mode target | grep -E \"iqn\\.[0-9]{4}-[0-9]{2}\" | grep Target "),
	)
	cmd.Stdout = &stdout

	if err := cmd.Run(); err != nil {
		log.Info(fmt.Sprintf(string(stderr.Bytes())))
		return 0
	}

	return _findMax(string(stdout.Bytes()))
}

func queryTargetId(iqn string) string {
	var stdout, stderr bytes.Buffer

	cmd := exec.Command("/bin/sh", "-c",
		fmt.Sprintf("tgtadm --lld iscsi --op show --mode target | grep -E %s", iqn),
	)
	cmd.Stdout = &stdout

	if err := cmd.Run(); err != nil {
		log.Info(fmt.Sprintf(string(stderr.Bytes())))
		return "-1"
	}

	return findTid(string(stdout.Bytes()))
}

func validateIQN(iqn string) bool {
	r, _ := regexp.Compile("iqn\\.(\\d{4}-\\d{2})\\.([^:]+)(:)([^,:\\s']+)")
	return r.MatchString(iqn)
}

func _findMax(s string) int {

	scanner := bufio.NewScanner(bufio.NewReader(strings.NewReader(s)))

	aa := []int{}
	for scanner.Scan() {
		line := scanner.Text()
		a, _ := strconv.Atoi(strings.Split(strings.Split(line, ":")[0], " ")[1])
		aa = append(aa, a)
	}

	if len(aa) == 0 {
		return 0
	}
	sort.Ints(aa)

	return aa[len(aa)-1]
}

func findTid(tartetIQN string) string {

	a := strings.Split(tartetIQN, ":")
	if len(a) == 0 {
		return "-1"
	}

	b := strings.Split(a[0], " ")

	if len(b) != 2 {
		return "-1"
	}

	return b[1]
}
