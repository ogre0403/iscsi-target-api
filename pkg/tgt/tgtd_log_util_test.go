package tgt

import (
	"github.com/stretchr/testify/assert"
	"regexp"
	"testing"
)

func Test_ValidateIQN(t *testing.T) {
	correct_iqn := "iqn.1993-08.org.debian:01:30d46bc47a7"
	b := validateIQN(correct_iqn)
	assert.Equal(t, true, b)

	wrong_iqn := "iq.2017-07.com.hiroom2:aaadd"
	b = validateIQN(wrong_iqn)
	assert.Equal(t, false, b)
}

func Test_FindMax(t *testing.T) {
	s := "Target 3: iqn.2017-07.com.hiroom2:aaadd\nTarget 7: iqn.2017-07.com.hiroom2:aaadd\nTarget 1: iqn.2017-07.com.hiroom2:aaadd"
	r := _findMax(s)
	assert.Equal(t, "7", r)
}

func Test_FindTd(t *testing.T) {
	s1 := "Target 3: iqn.2017-07.com.hiroom2:aaadd"
	tid1 := findTid(s1)
	assert.Equal(t, "3", tid1)

	s2 := ""
	tid2 := findTid(s2)
	assert.Equal(t, "-1", tid2)

}

func Test_SizeRegex(t *testing.T) {
	r, _ := regexp.Compile("[0-9]+m$")
	assert.Equal(t, false, r.MatchString("1024mm"))
	assert.Equal(t, true, r.MatchString("1024m"))
}
