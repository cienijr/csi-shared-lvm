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
			expectedArgs: strings.Fields("--noheadings --nosuffix --units b -o lv_name,lv_size,lv_attr,lv_tags test-vg/test-lv"),
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

func TestBuildLvremoveCmd(t *testing.T) {
	tests := []struct {
		name         string
		vg           string
		lv           string
		expectedCmd  string
		expectedArgs []string
	}{
		{
			name:         "should delete lv successfully",
			vg:           "test-vg",
			lv:           "test-lv",
			expectedCmd:  "lvremove",
			expectedArgs: strings.Fields("-f test-vg/test-lv"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd, args := buildLvremoveCmd(tt.vg, tt.lv)
			assert.Equal(t, tt.expectedCmd, cmd)
			assert.Equal(t, tt.expectedArgs, args)
		})
	}
}

func TestBuildLvextendCmd(t *testing.T) {
	tests := []struct {
		name         string
		vg           string
		lv           string
		size         int64
		expectedCmd  string
		expectedArgs []string
	}{
		{
			name:         "should resize lv successfully",
			vg:           "test-vg",
			lv:           "test-lv",
			size:         2147483648,
			expectedCmd:  "lvextend",
			expectedArgs: strings.Fields("-L 2147483648b test-vg/test-lv"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd, args := buildLvextendCmd(tt.vg, tt.lv, tt.size)
			assert.Equal(t, tt.expectedCmd, cmd)
			assert.Equal(t, tt.expectedArgs, args)
		})
	}
}

func TestBuildLvchangeActivateCmd(t *testing.T) {
	tests := []struct {
		name         string
		vg           string
		lv           string
		expectedCmd  string
		expectedArgs []string
	}{
		{
			name:         "should activate lv successfully",
			vg:           "test-vg",
			lv:           "test-lv",
			expectedCmd:  "lvchange",
			expectedArgs: strings.Fields("-ay test-vg/test-lv"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd, args := buildLvchangeActivateCmd(tt.vg, tt.lv)
			assert.Equal(t, tt.expectedCmd, cmd)
			assert.Equal(t, tt.expectedArgs, args)
		})
	}
}

func TestBuildLvchangeDeactivateCmd(t *testing.T) {
	tests := []struct {
		name         string
		vg           string
		lv           string
		expectedCmd  string
		expectedArgs []string
	}{
		{
			name:         "should deactivate lv successfully",
			vg:           "test-vg",
			lv:           "test-lv",
			expectedCmd:  "lvchange",
			expectedArgs: strings.Fields("-an test-vg/test-lv"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd, args := buildLvchangeDeactivateCmd(tt.vg, tt.lv)
			assert.Equal(t, tt.expectedCmd, cmd)
			assert.Equal(t, tt.expectedArgs, args)
		})
	}
}

func TestBuildVgsCmg(t *testing.T) {
	tests := []struct {
		name         string
		vg           string
		expectedCmd  string
		expectedArgs []string
	}{
		{
			name:         "should get vg successfully",
			vg:           "test-vg",
			expectedCmd:  "vgs",
			expectedArgs: strings.Fields("--noheadings --nosuffix --units b -o vg_name,vg_free test-vg"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd, args := buildVgsCmg(tt.vg)
			assert.Equal(t, tt.expectedCmd, cmd)
			assert.Equal(t, tt.expectedArgs, args)
		})
	}
}
