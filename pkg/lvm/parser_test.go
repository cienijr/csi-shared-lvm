package lvm

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

type mockExitError struct {
	exitCode int
}

func (m *mockExitError) Error() string {
	return fmt.Sprintf("mock exit code %d", m.exitCode)
}

func (m *mockExitError) ExitCode() int {
	return m.exitCode
}

func TestParseLVSOutput(t *testing.T) {
	tests := []struct {
		name        string
		vg          string
		stdout      string
		stderr      string
		err         error
		expectedLV  *LogicalVolume
		expectedErr error
	}{
		{
			name:   "should parse lvs output successfully",
			vg:     "test-vg",
			stdout: "  test-lv 1073741824B -wi-a----- test-tag",
			expectedLV: &LogicalVolume{
				Name: "test-lv",
				VG:   "test-vg",
				Size: 1073741824,
				Tags: []string{"test-tag"},
				Attr: "-wi-a-----",
			},
		},
		{
			name:   "should parse lvs output successfully with multiple tags",
			vg:     "test-vg",
			stdout: "  test-lv 1073741824B -wi------- test-tag,test-tag2,test-tag3",
			expectedLV: &LogicalVolume{
				Name: "test-lv",
				VG:   "test-vg",
				Size: 1073741824,
				Tags: []string{"test-tag", "test-tag2", "test-tag3"},
				Attr: "-wi-------",
			},
		},
		{
			name:   "should parse lvs output successfully with no tags",
			vg:     "test-vg",
			stdout: "  test-lv 1073741824B -wi-ao----",
			expectedLV: &LogicalVolume{
				Name: "test-lv",
				VG:   "test-vg",
				Size: 1073741824,
				Attr: "-wi-ao----",
			},
		},
		{
			name:        "should return nil if lv not found",
			vg:          "test-vg",
			stdout:      "",
			stderr:      `  Failed to find logical volume "test-vg/test-lv"`,
			err:         &mockExitError{exitCode: 5},
			expectedErr: nil,
		},
		{
			name:        "should return nil if vg not found",
			vg:          "test-vg",
			stdout:      "",
			stderr:      `  Volume group "test-vg" not found`,
			err:         &mockExitError{exitCode: 5},
			expectedErr: nil,
		},
		{
			name:        "should return error if command fails",
			vg:          "test-vg",
			stdout:      "",
			stderr:      "some error output",
			err:         fmt.Errorf("some error"),
			expectedErr: fmt.Errorf("failed to get lv: some error, stderr: some error output"),
		},
		{
			name:        "should return error on malformed output",
			vg:          "test-vg",
			stdout:      "malformed",
			expectedErr: fmt.Errorf("failed to parse lvs output: malformed"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lv, err := parseLvsOutput(tt.vg, tt.stdout, tt.stderr, tt.err)
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

func TestParseAttr(t *testing.T) {
	tests := []struct {
		name     string
		attrs    []Attr
		isActive bool
	}{
		{
			name:     "should parse isActive true",
			attrs:    []Attr{"-wi-a-----", "-wi-ao----", "----a-----"}, // 5th bit = a
			isActive: true,
		},
		{
			name:     "should parse isActive false",
			attrs:    []Attr{"-wi-------", "-wi-h----", "-w--s-----"}, // 5th bit != a
			isActive: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for attr := range tt.attrs {
				assert.Equal(t, tt.isActive, tt.attrs[attr].IsActive())
			}
		})
	}
}
