package driver

import (
	"context"
	"fmt"
	"testing"

	"github.com/container-storage-interface/spec/lib/go/csi"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/cienijr/csi-shared-lvm/pkg/lvm"
)

func TestNodeStageVolume(t *testing.T) {
	tests := []struct {
		name        string
		req         *csi.NodeStageVolumeRequest
		mockLVM     *mockLVM
		expectedErr codes.Code
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
			expectedErr: codes.OK,
		},
		{
			name: "should return success if volume already staged",
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
			expectedErr: codes.OK,
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
			expectedErr: codes.Internal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			driver := NewDriver("test-endpoint", nil, tt.mockLVM)
			_, err := driver.NodeStageVolume(context.Background(), tt.req)
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
