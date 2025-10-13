package lvm

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBuildLvcreateCmd(t *testing.T) {
	tests := []struct {
		name         string
		vg           string
		lv           string
		size         int64
		tags         []string
		expectedCmd  string
		expectedArgs []string
	}{
		{
			name:         "should create lv with correct command",
			vg:           "test-vg",
			lv:           "test-lv",
			size:         1024 * 1024 * 1024,
			tags:         []string{"test-tag"},
			expectedCmd:  "lvcreate",
			expectedArgs: strings.Fields("--name test-lv --wipesignatures y --yes --size 1073741824b --setautoactivation n --addtag test-tag test-vg"),
		},
		{
			name:         "should create lv without tags",
			vg:           "test-vg",
			lv:           "test-lv",
			size:         1024 * 1024 * 1024,
			tags:         []string{},
			expectedCmd:  "lvcreate",
			expectedArgs: strings.Fields("--name test-lv --wipesignatures y --yes --size 1073741824b --setautoactivation n test-vg"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd, args := buildLvcreateCmd(tt.vg, tt.lv, tt.size, tt.tags)
			assert.Equal(t, tt.expectedCmd, cmd)
			assert.Equal(t, tt.expectedArgs, args)
		})
	}
}

func TestBuildLvsCmd(t *testing.T) {
	tests := []struct {
		name         string
		vg           string
		lv           string
		expectedCmd  string
		expectedArgs []string
	}{
		{
			name:         "should get lv successfully",
			vg:           "test-vg",
			lv:           "test-lv",
			expectedCmd:  "lvs",
			expectedArgs: strings.Fields("--noheadings --nosuffix --units b -o lv_name,lv_size,lv_tags test-vg/test-lv"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd, args := buildLvsCmd(tt.vg, tt.lv)
			assert.Equal(t, tt.expectedCmd, cmd)
			assert.Equal(t, tt.expectedArgs, args)
		})
	}
}
