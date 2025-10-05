package lvm

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

type mockExecutor struct {
	run func(command string, args ...string) (stdout []byte, stderr []byte, err error)
}

func (m *mockExecutor) Run(command string, args ...string) (stdout []byte, stderr []byte, err error) {
	return m.run(command, args...)
}

type mockExitError struct {
	exitCode int
}

func (m *mockExitError) Error() string {
	return fmt.Sprintf("mock exit code %d", m.exitCode)
}

func (m *mockExitError) ExitCode() int {
	return m.exitCode
}

func TestCreateLV(t *testing.T) {
	tests := []struct {
		name        string
		vg          string
		lv          string
		size        int64
		tags        []string
		expectedCmd string
	}{
		{
			name:        "should create lv with correct command",
			vg:          "test-vg",
			lv:          "test-lv",
			size:        1024 * 1024 * 1024,
			tags:        []string{"test-tag"},
			expectedCmd: "lvcreate --name test-lv --wipesignatures y --yes --size 1073741824b --setautoactivation n --addtag test-tag test-vg",
		},
		{
			name:        "should create lv without tags",
			vg:          "test-vg",
			lv:          "test-lv",
			size:        1024 * 1024 * 1024,
			tags:        []string{},
			expectedCmd: "lvcreate --name test-lv --wipesignatures y --yes --size 1073741824b --setautoactivation n test-vg",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockExec := &mockExecutor{
				run: func(command string, args ...string) (stdout []byte, stderr []byte, err error) {
					cmd := command + " " + strings.Join(args, " ")
					assert.Equal(t, tt.expectedCmd, cmd)
					return nil, nil, nil
				},
			}
			lvmClient := NewLVM(mockExec)
			err := lvmClient.CreateLV(tt.vg, tt.lv, tt.size, tt.tags)
			assert.NoError(t, err)
		})
	}
}

func TestGetLV(t *testing.T) {
	tests := []struct {
		name        string
		vg          string
		lv          string
		mockStdout  string
		mockStderr  string
		expectedCmd string
		expectedLV  *LogicalVolume
		mockErr     error
		expectedErr error
	}{
		{
			name:        "should get lv successfully",
			vg:          "test-vg",
			lv:          "test-lv",
			mockStdout:  "  test-lv 1073741824B test-tag",
			expectedCmd: "lvs --noheadings --nosuffix --units b -o lv_name,lv_size,lv_tags test-vg/test-lv",
			expectedLV: &LogicalVolume{
				Name: "test-lv",
				VG:   "test-vg",
				Size: 1073741824,
				Tags: []string{"test-tag"},
			},
		},
		{
			name:        "should return nil if lv not found",
			vg:          "test-vg",
			lv:          "test-lv",
			mockStdout:  "",
			expectedCmd: "lvs --noheadings --nosuffix --units b -o lv_name,lv_size,lv_tags test-vg/test-lv",
			expectedLV:  nil,
		},
		{
			name:        "should return error if command fails",
			vg:          "test-vg",
			lv:          "test-lv",
			mockStdout:  "",
			expectedCmd: "lvs --noheadings --nosuffix --units b -o lv_name,lv_size,lv_tags test-vg/test-lv",
			mockErr:     fmt.Errorf("some error"),
			mockStderr:  "some output",
			expectedErr: fmt.Errorf("failed to get lv: some error, stderr: some output"),
		},
		{
			name:        "should not return error if command fails for non existing lv",
			vg:          "test-vg",
			lv:          "test-lv",
			mockStdout:  "",
			expectedCmd: "lvs --noheadings --nosuffix --units b -o lv_name,lv_size,lv_tags test-vg/test-lv",
			mockErr:     &mockExitError{exitCode: 5},
			mockStderr:  `  Failed to find logical volume "test-vg/test-lv"`,
		},
		{
			name:        "should not return error if command fails for non existing vg",
			vg:          "test-vg",
			lv:          "test-lv",
			mockStdout:  "",
			expectedCmd: "lvs --noheadings --nosuffix --units b -o lv_name,lv_size,lv_tags test-vg/test-lv",
			mockErr:     &mockExitError{exitCode: 5},
			mockStderr:  `  Volume group "test-vg" not found`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockExec := &mockExecutor{
				run: func(command string, args ...string) (stdout []byte, stderr []byte, err error) {
					cmd := command + " " + strings.Join(args, " ")
					assert.Equal(t, tt.expectedCmd, cmd)
					return []byte(tt.mockStdout), []byte(tt.mockStderr), tt.mockErr
				},
			}
			lvmClient := NewLVM(mockExec)
			lv, err := lvmClient.GetLV(tt.vg, tt.lv)
			if tt.expectedErr != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedErr, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedLV, lv)
			}
		})
	}
}
