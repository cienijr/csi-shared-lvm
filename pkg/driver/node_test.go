package driver

import (
	"context"
	"fmt"
	"testing"

	"github.com/container-storage-interface/spec/lib/go/csi"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"k8s.io/mount-utils"
	testingexec "k8s.io/utils/exec/testing"

	"github.com/cienijr/csi-shared-lvm/pkg/lvm"
)

func TestNodeStageVolume(t *testing.T) {
	tests := []struct {
		name        string
		req         *csi.NodeStageVolumeRequest
		mockLVM     *mockLVM
		mounter     *mount.FakeMounter
		actions     []testingexec.FakeCommandAction
		expectedErr codes.Code
		expectedLog []mount.FakeAction
	}{
		{
			name: "should stage volume successfully",
			req: &csi.NodeStageVolumeRequest{
				VolumeId:          "test-vg/test-lv",
				StagingTargetPath: "/test/path",
				VolumeCapability: &csi.VolumeCapability{
					AccessMode: &csi.VolumeCapability_AccessMode{
						Mode: csi.VolumeCapability_AccessMode_SINGLE_NODE_WRITER,
					},
				},
			},
			mockLVM: &mockLVM{
				getLV: func(vg, name string) (*lvm.LogicalVolume, error) {
					return &lvm.LogicalVolume{
						Name: "test-lv",
						VG:   "test-vg",
						Attr: "-wi-------",
					}, nil
				},
				activateLV: func(vg, name string) error {
					return nil
				},
			},
			mounter:     &mount.FakeMounter{},
			actions:     []testingexec.FakeCommandAction{},
			expectedErr: codes.OK,
			expectedLog: []mount.FakeAction{
				{
					Action: "mount",
					Source: "/dev/test-vg/test-lv",
					Target: "/test/path",
					FSType: "ext4",
				},
			},
		},
		{
			name: "should stage volume successfully when already activated but not mounted",
			req: &csi.NodeStageVolumeRequest{
				VolumeId:          "test-vg/test-lv",
				StagingTargetPath: "/test/path",
				VolumeCapability: &csi.VolumeCapability{
					AccessMode: &csi.VolumeCapability_AccessMode{
						Mode: csi.VolumeCapability_AccessMode_SINGLE_NODE_WRITER,
					},
				},
			},
			mockLVM: &mockLVM{
				getLV: func(vg, name string) (*lvm.LogicalVolume, error) {
					return &lvm.LogicalVolume{
						Name: "test-lv",
						VG:   "test-vg",
						Attr: "-wi-a-----",
					}, nil
				},
				activateLV: func(vg, name string) error {
					assert.Fail(t, "activateLV should not be called")
					return nil
				},
			},
			mounter:     &mount.FakeMounter{},
			actions:     []testingexec.FakeCommandAction{},
			expectedErr: codes.OK,
			expectedLog: []mount.FakeAction{
				{
					Action: "mount",
					Source: "/dev/test-vg/test-lv",
					Target: "/test/path",
					FSType: "ext4",
				},
			},
		},
		{
			name: "should stage block volume successfully",
			req: &csi.NodeStageVolumeRequest{
				VolumeId:          "test-vg/test-lv",
				StagingTargetPath: "/test/path",
				VolumeCapability: &csi.VolumeCapability{
					AccessType: &csi.VolumeCapability_Block{
						Block: &csi.VolumeCapability_BlockVolume{},
					},
				},
			},
			mockLVM: &mockLVM{
				getLV: func(vg, name string) (*lvm.LogicalVolume, error) {
					return &lvm.LogicalVolume{
						Name: "test-lv",
						VG:   "test-vg",
						Attr: "-wi-------",
					}, nil
				},
				activateLV: func(vg, name string) error {
					return nil
				},
			},
			mounter:     &mount.FakeMounter{},
			expectedErr: codes.OK,
		},
		{
			name: "should return success if filesystem volume is already activated",
			req: &csi.NodeStageVolumeRequest{
				VolumeId:          "test-vg/test-lv",
				StagingTargetPath: "/test/path",
				VolumeCapability: &csi.VolumeCapability{
					AccessMode: &csi.VolumeCapability_AccessMode{
						Mode: csi.VolumeCapability_AccessMode_SINGLE_NODE_WRITER,
					},
				},
			},
			mockLVM: &mockLVM{
				getLV: func(vg, name string) (*lvm.LogicalVolume, error) {
					return &lvm.LogicalVolume{
						Name: "test-lv",
						VG:   "test-vg",
						Attr: "-wi-a-----",
					}, nil
				},
				activateLV: func(vg, name string) error {
					assert.Fail(t, "activateLV should not have been called")
					return nil
				},
			},
			mounter:     &mount.FakeMounter{},
			expectedErr: codes.OK,
			expectedLog: []mount.FakeAction{
				{
					Action: "mount",
					Source: "/dev/test-vg/test-lv",
					Target: "/test/path",
					FSType: "ext4",
				},
			},
		},
		{
			name: "should fail if volume not found",
			req: &csi.NodeStageVolumeRequest{
				VolumeId:          "test-vg/test-lv",
				StagingTargetPath: "/test/path",
				VolumeCapability: &csi.VolumeCapability{
					AccessMode: &csi.VolumeCapability_AccessMode{
						Mode: csi.VolumeCapability_AccessMode_SINGLE_NODE_WRITER,
					},
				},
			},
			mockLVM: &mockLVM{
				getLV: func(vg, name string) (*lvm.LogicalVolume, error) {
					return nil, nil
				},
				activateLV: func(vg, name string) error {
					assert.Fail(t, "activateLV should not have been called")
					return nil
				},
			},
			mounter:     &mount.FakeMounter{},
			expectedErr: codes.NotFound,
		},
		{
			name: "should fail on internal error on get lv",
			req: &csi.NodeStageVolumeRequest{
				VolumeId:          "test-vg/test-lv",
				StagingTargetPath: "/test/path",
				VolumeCapability: &csi.VolumeCapability{
					AccessMode: &csi.VolumeCapability_AccessMode{
						Mode: csi.VolumeCapability_AccessMode_SINGLE_NODE_WRITER,
					},
				},
			},
			mockLVM: &mockLVM{
				getLV: func(vg, name string) (*lvm.LogicalVolume, error) {
					return nil, fmt.Errorf("some error")
				},
				activateLV: func(vg, name string) error {
					assert.Fail(t, "activateLV should not have been called")
					return nil
				},
			},
			mounter:     &mount.FakeMounter{},
			expectedErr: codes.Internal,
		},
		{
			name: "should fail on internal error on activate lv",
			req: &csi.NodeStageVolumeRequest{
				VolumeId:          "test-vg/test-lv",
				StagingTargetPath: "/test/path",
				VolumeCapability: &csi.VolumeCapability{
					AccessMode: &csi.VolumeCapability_AccessMode{
						Mode: csi.VolumeCapability_AccessMode_SINGLE_NODE_WRITER,
					},
				},
			},
			mockLVM: &mockLVM{
				getLV: func(vg, name string) (*lvm.LogicalVolume, error) {
					return &lvm.LogicalVolume{
						Name: "test-lv",
						VG:   "test-vg",
						Attr: "-wi-------",
					}, nil
				},
				activateLV: func(vg, name string) error {
					return fmt.Errorf("some other error")
				},
			},
			mounter:     &mount.FakeMounter{},
			expectedErr: codes.Internal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			driver := NewDriver("test-endpoint", nil, tt.mockLVM)
			exec := &testingexec.FakeExec{CommandScript: tt.actions}
			driver.mounter = &mount.SafeFormatAndMount{Interface: tt.mounter, Exec: exec}
			_, err := driver.NodeStageVolume(context.Background(), tt.req)
			if tt.expectedErr == codes.OK {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
				st, ok := status.FromError(err)
				assert.True(t, ok)
				assert.Equal(t, tt.expectedErr, st.Code())
			}
			assert.Equal(t, tt.expectedLog, tt.mounter.GetLog())
		})
	}
}

func TestNodeUnstageVolume(t *testing.T) {
	tests := []struct {
		name        string
		req         *csi.NodeUnstageVolumeRequest
		mockLVM     *mockLVM
		mounter     *mount.FakeMounter
		expectedErr codes.Code
		expectedLog []mount.FakeAction
	}{
		{
			name: "should unstage volume successfully",
			req: &csi.NodeUnstageVolumeRequest{
				VolumeId:          "test-vg/test-lv",
				StagingTargetPath: "/test/path",
			},
			mockLVM: &mockLVM{
				getLV: func(vg, name string) (*lvm.LogicalVolume, error) {
					return &lvm.LogicalVolume{
						Name: "test-lv",
						VG:   "test-vg",
						Attr: "-wi-a-----",
					}, nil
				},
				deactivateLV: func(vg, name string) error {
					return nil
				},
			},
			mounter: &mount.FakeMounter{
				MountPoints: []mount.MountPoint{
					{
						Device: "/dev/test-vg/test-lv",
						Path:   "/test/path",
					},
				},
			},
			expectedErr: codes.OK,
			//expectedLog: []mount.FakeAction{
			//	{
			//		Action: "unmount",
			//		Source: "/dev/test-vg/test-lv",
			//		Target: "/test/path",
			//		FSType: "ext4",
			//	},
			//},
		},
		{
			name: "should unstage volume successfully (block volume)",
			req: &csi.NodeUnstageVolumeRequest{
				VolumeId:          "test-vg/test-lv",
				StagingTargetPath: "/test/path",
			},
			mockLVM: &mockLVM{
				getLV: func(vg, name string) (*lvm.LogicalVolume, error) {
					return &lvm.LogicalVolume{
						Name: "test-lv",
						VG:   "test-vg",
						Attr: "-wi-a-----",
					}, nil
				},
				deactivateLV: func(vg, name string) error {
					return nil
				},
			},
			mounter:     &mount.FakeMounter{},
			expectedErr: codes.OK,
		},
		{
			name: "should return success if volume already unstaged",
			req: &csi.NodeUnstageVolumeRequest{
				VolumeId:          "test-vg/test-lv",
				StagingTargetPath: "/test/path",
			},
			mockLVM: &mockLVM{
				getLV: func(vg, name string) (*lvm.LogicalVolume, error) {
					return &lvm.LogicalVolume{
						Name: "test-lv",
						VG:   "test-vg",
						Attr: "-wi-------",
					}, nil
				},
				deactivateLV: func(vg, name string) error {
					assert.Fail(t, "deactivateLV should not have been called")
					return nil
				},
			},
			mounter:     &mount.FakeMounter{},
			expectedErr: codes.OK,
		},
		{
			name: "should return success if volume not found",
			req: &csi.NodeUnstageVolumeRequest{
				VolumeId:          "test-vg/test-lv",
				StagingTargetPath: "/test/path",
			},
			mockLVM: &mockLVM{
				getLV: func(vg, name string) (*lvm.LogicalVolume, error) {
					return nil, nil
				},
				deactivateLV: func(vg, name string) error {
					assert.Fail(t, "deactivateLV should not have been called")
					return nil
				},
			},
			mounter:     &mount.FakeMounter{},
			expectedErr: codes.OK,
		},
		{
			name: "should fail on internal error on get lv",
			req: &csi.NodeUnstageVolumeRequest{
				VolumeId:          "test-vg/test-lv",
				StagingTargetPath: "/test/path",
			},
			mockLVM: &mockLVM{
				getLV: func(vg, name string) (*lvm.LogicalVolume, error) {
					return nil, fmt.Errorf("some other error")
				},
				deactivateLV: func(vg, name string) error {
					assert.Fail(t, "deactivateLV should not have been called")
					return nil
				},
			},
			mounter: &mount.FakeMounter{
				MountPoints: []mount.MountPoint{
					{
						Device: "/dev/test-vg/test-lv",
						Path:   "/test/path",
					},
				},
			},
			expectedErr: codes.Internal,
		},
		{
			name: "should fail on internal error on deactivate lv",
			req: &csi.NodeUnstageVolumeRequest{
				VolumeId:          "test-vg/test-lv",
				StagingTargetPath: "/test/path",
			},
			mockLVM: &mockLVM{
				getLV: func(vg, name string) (*lvm.LogicalVolume, error) {
					return &lvm.LogicalVolume{
						Name: "test-lv",
						VG:   "test-vg",
						Attr: "-wi-a-----",
					}, nil
				},
				deactivateLV: func(vg, name string) error {
					return fmt.Errorf("some other error")
				},
			},
			mounter: &mount.FakeMounter{
				MountPoints: []mount.MountPoint{
					{
						Device: "/dev/test-vg/test-lv",
						Path:   "/test/path",
					},
				},
			},
			expectedErr: codes.Internal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			driver := NewDriver("test-endpoint", nil, tt.mockLVM)
			driver.mounter = &mount.SafeFormatAndMount{Interface: tt.mounter, Exec: &testingexec.FakeExec{}}
			_, err := driver.NodeUnstageVolume(context.Background(), tt.req)
			if tt.expectedErr == codes.OK {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
				st, ok := status.FromError(err)
				assert.True(t, ok)
				assert.Equal(t, tt.expectedErr, st.Code())
			}
			assert.Equal(t, tt.expectedLog, tt.mounter.GetLog())
		})
	}
}
