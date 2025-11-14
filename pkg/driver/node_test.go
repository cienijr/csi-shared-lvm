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

type mockResizer struct {
	needResize func(devicePath, deviceMountPath string) (bool, error)
	resize     func(devicePath, deviceMountPath string) (bool, error)
}

func (m *mockResizer) NeedResize(devicePath, deviceMountPath string) (bool, error) {
	if m.needResize != nil {
		return m.needResize(devicePath, deviceMountPath)
	}
	return false, nil
}

func (m *mockResizer) Resize(devicePath, deviceMountPath string) (bool, error) {
	if m.resize != nil {
		return m.resize(devicePath, deviceMountPath)
	}
	return true, nil
}

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
			driver.resizer = &mockResizer{
				needResize: func(devicePath, deviceMountPath string) (bool, error) {
					return false, nil
				},
				resize: func(devicePath, deviceMountPath string) (bool, error) {
					assert.Fail(t, "resize should not have been called")
					return false, nil
				},
			}
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

func TestNodeExpandVolume(t *testing.T) {
	tests := []struct {
		name        string
		req         *csi.NodeExpandVolumeRequest
		mounter     *mount.FakeMounter
		mockResizer *mockResizer
		useTempDir  bool
		expectedErr codes.Code
	}{
		{
			name: "should expand volume successfully",
			req: &csi.NodeExpandVolumeRequest{
				VolumeId:   "test-vg/test-lv",
				VolumePath: "/test/path",
			},
			mounter: &mount.FakeMounter{
				MountPoints: []mount.MountPoint{
					{
						Device: "/dev/test-vg/test-lv",
						Path:   "/test/path",
					},
				},
			},
			mockResizer: &mockResizer{
				resize: func(devicePath, deviceMountPath string) (bool, error) {
					return true, nil
				},
			},
			useTempDir:  true,
			expectedErr: codes.OK,
		},
		{
			name: "should ignore block volumes (not mount point)",
			req: &csi.NodeExpandVolumeRequest{
				VolumeId:   "test-vg/test-lv",
				VolumePath: "/dev/test-vg/test-lv",
				VolumeCapability: &csi.VolumeCapability{
					AccessType: &csi.VolumeCapability_Block{
						Block: &csi.VolumeCapability_BlockVolume{},
					},
				},
			},
			mounter:     &mount.FakeMounter{},
			expectedErr: codes.OK,
		},
		{
			name: "should fail if volume id is missing",
			req: &csi.NodeExpandVolumeRequest{
				VolumePath: "/test/path",
			},
			mounter:     &mount.FakeMounter{},
			expectedErr: codes.InvalidArgument,
		},
		{
			name: "should fail if volume path is missing",
			req: &csi.NodeExpandVolumeRequest{
				VolumeId: "test-vg/test-lv",
			},
			mounter:     &mount.FakeMounter{},
			expectedErr: codes.InvalidArgument,
		},
		{
			name: "should fail if volume id format is invalid",
			req: &csi.NodeExpandVolumeRequest{
				VolumeId:   "invalid",
				VolumePath: "/test/path",
			},
			mounter:     &mount.FakeMounter{},
			expectedErr: codes.InvalidArgument,
		},
		{
			name: "should fail if resize fails",
			req: &csi.NodeExpandVolumeRequest{
				VolumeId:   "test-vg/test-lv",
				VolumePath: "/test/path",
			},
			mounter: &mount.FakeMounter{
				MountPoints: []mount.MountPoint{
					{
						Device: "/dev/test-vg/test-lv",
						Path:   "/test/path",
					},
				},
			},
			mockResizer: &mockResizer{
				resize: func(devicePath, deviceMountPath string) (bool, error) {
					return false, fmt.Errorf("resize failed")
				},
			},
			useTempDir:  true,
			expectedErr: codes.Internal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.useTempDir {
				tmpDir := t.TempDir()
				tt.req.VolumePath = tmpDir
				if tt.mounter != nil {
					for i := range tt.mounter.MountPoints {
						if tt.mounter.MountPoints[i].Path == "/test/path" {
							tt.mounter.MountPoints[i].Path = tmpDir
						}
					}
				}
			}

			driver := NewDriver("test-endpoint", nil, nil)
			driver.mounter = &mount.SafeFormatAndMount{Interface: tt.mounter, Exec: &testingexec.FakeExec{}}
			driver.resizer = &mockResizer{}
			if tt.mockResizer != nil {
				driver.resizer = tt.mockResizer
			}

			_, err := driver.NodeExpandVolume(context.Background(), tt.req)
			if tt.expectedErr == codes.OK {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
				st, ok := status.FromError(err)
				assert.True(t, ok)
				assert.Equal(t, tt.expectedErr, st.Code())
			}
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
