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

func TestCreateVolume(t *testing.T) {
	tests := []struct {
		name        string
		req         *csi.CreateVolumeRequest
		allowedVGs  []string
		mockLVM     *mockLVM
		expectedErr codes.Code
	}{
		{
			name: "should create volume successfully",
			req: &csi.CreateVolumeRequest{
				Name: "test-lv",
				CapacityRange: &csi.CapacityRange{
					RequiredBytes: 1024 * 1024 * 1024,
				},
				VolumeCapabilities: []*csi.VolumeCapability{
					{
						AccessType: &csi.VolumeCapability_Mount{
							Mount: &csi.VolumeCapability_MountVolume{},
						},
						AccessMode: &csi.VolumeCapability_AccessMode{
							Mode: csi.VolumeCapability_AccessMode_SINGLE_NODE_WRITER,
						},
					},
				},
				Parameters: map[string]string{
					volumeGroupKey: "test-vg",
				},
			},
			mockLVM: func() *mockLVM {
				var getLV *lvm.LogicalVolume

				return &mockLVM{
					getLV: func(vg, name string) (*lvm.LogicalVolume, error) {
						// first call returns nil, after createLV returns the mock
						return getLV, nil
					},
					createLV: func(vg, name string, size int64, tags []string) error {
						getLV = &lvm.LogicalVolume{
							Name: name,
							VG:   vg,
							Size: size,
							Tags: tags,
						}

						return nil
					},
				}
			}(),
			expectedErr: codes.OK,
		},
		{
			name: "should return success if volume already exists with correct size",
			req: &csi.CreateVolumeRequest{
				Name: "test-lv",
				CapacityRange: &csi.CapacityRange{
					RequiredBytes: 1024 * 1024 * 1024,
				},
				VolumeCapabilities: []*csi.VolumeCapability{
					{
						AccessType: &csi.VolumeCapability_Mount{
							Mount: &csi.VolumeCapability_MountVolume{},
						},
						AccessMode: &csi.VolumeCapability_AccessMode{
							Mode: csi.VolumeCapability_AccessMode_SINGLE_NODE_WRITER,
						},
					},
				},
				Parameters: map[string]string{
					volumeGroupKey: "test-vg",
				},
			},
			mockLVM: &mockLVM{
				getLV: func(vg, name string) (*lvm.LogicalVolume, error) {
					return &lvm.LogicalVolume{
						Name: "test-lv",
						VG:   "test-vg",
						Size: 1024 * 1024 * 1024,
					}, nil
				},
				createLV: func(vg, name string, size int64, tags []string) error {
					assert.Fail(t, "createLV should not have been called")
					return nil
				},
			},
			expectedErr: codes.OK,
		},
		{
			name: "should return success if volume already exists with greater size",
			req: &csi.CreateVolumeRequest{
				Name: "test-lv",
				CapacityRange: &csi.CapacityRange{
					RequiredBytes: 1024 * 1024 * 1024,
				},
				VolumeCapabilities: []*csi.VolumeCapability{
					{
						AccessType: &csi.VolumeCapability_Mount{
							Mount: &csi.VolumeCapability_MountVolume{},
						},
						AccessMode: &csi.VolumeCapability_AccessMode{
							Mode: csi.VolumeCapability_AccessMode_SINGLE_NODE_WRITER,
						},
					},
				},
				Parameters: map[string]string{
					volumeGroupKey: "test-vg",
				},
			},
			mockLVM: &mockLVM{
				getLV: func(vg, name string) (*lvm.LogicalVolume, error) {
					return &lvm.LogicalVolume{
						Name: "test-lv",
						VG:   "test-vg",
						Size: 10 * 1024 * 1024 * 1024,
					}, nil
				},
				createLV: func(vg, name string, size int64, tags []string) error {
					assert.Fail(t, "createLV should not have been called")
					return nil
				},
			},
			expectedErr: codes.OK,
		},
		{
			name: "should fail if volume already exists with smaller size",
			req: &csi.CreateVolumeRequest{
				Name: "test-lv",
				CapacityRange: &csi.CapacityRange{
					RequiredBytes: 2 * 1024 * 1024 * 1024,
				},
				VolumeCapabilities: []*csi.VolumeCapability{
					{
						AccessType: &csi.VolumeCapability_Mount{
							Mount: &csi.VolumeCapability_MountVolume{},
						},
						AccessMode: &csi.VolumeCapability_AccessMode{
							Mode: csi.VolumeCapability_AccessMode_SINGLE_NODE_WRITER,
						},
					},
				},
				Parameters: map[string]string{
					volumeGroupKey: "test-vg",
				},
			},
			mockLVM: &mockLVM{
				getLV: func(vg, name string) (*lvm.LogicalVolume, error) {
					return &lvm.LogicalVolume{
						Name: "test-lv",
						VG:   "test-vg",
						Size: 1024 * 1024 * 1024,
					}, nil
				},
				createLV: func(vg, name string, size int64, tags []string) error {
					assert.Fail(t, "createLV should not have been called")
					return nil
				},
			},
			expectedErr: codes.AlreadyExists,
		},
		{
			name: "should fail if volume group is not allowed",
			req: &csi.CreateVolumeRequest{
				Name: "test-lv",
				CapacityRange: &csi.CapacityRange{
					RequiredBytes: 1024 * 1024 * 1024,
				},
				VolumeCapabilities: []*csi.VolumeCapability{
					{
						AccessType: &csi.VolumeCapability_Mount{
							Mount: &csi.VolumeCapability_MountVolume{},
						},
						AccessMode: &csi.VolumeCapability_AccessMode{
							Mode: csi.VolumeCapability_AccessMode_SINGLE_NODE_WRITER,
						},
					},
				},
				Parameters: map[string]string{
					volumeGroupKey: "not-allowed-vg",
				},
			},
			allowedVGs: []string{"test-vg"},
			mockLVM: &mockLVM{
				getLV: func(vg, name string) (*lvm.LogicalVolume, error) {
					return nil, nil
				},
				createLV: func(vg, name string, size int64, tags []string) error {
					assert.Fail(t, "createLV should not have been called")
					return nil
				},
			},
			expectedErr: codes.InvalidArgument,
		},
		{
			name: "should fail if get lv fails",
			req: &csi.CreateVolumeRequest{
				Name: "test-lv",
				CapacityRange: &csi.CapacityRange{
					RequiredBytes: 1024 * 1024 * 1024,
				},
				VolumeCapabilities: []*csi.VolumeCapability{
					{
						AccessType: &csi.VolumeCapability_Mount{
							Mount: &csi.VolumeCapability_MountVolume{},
						},
						AccessMode: &csi.VolumeCapability_AccessMode{
							Mode: csi.VolumeCapability_AccessMode_SINGLE_NODE_WRITER,
						},
					},
				},
				Parameters: map[string]string{
					volumeGroupKey: "test-vg",
				},
			},
			mockLVM: &mockLVM{
				getLV: func(vg, name string) (*lvm.LogicalVolume, error) {
					return nil, fmt.Errorf("some error")
				},
				createLV: func(vg, name string, size int64, tags []string) error {
					assert.Fail(t, "createLV should not have been called")
					return nil
				},
			},
			expectedErr: codes.Internal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			driver := NewDriver("test-endpoint", tt.allowedVGs, tt.mockLVM)
			_, err := driver.CreateVolume(context.Background(), tt.req)
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

func TestDeleteVolume(t *testing.T) {
	tests := []struct {
		name        string
		req         *csi.DeleteVolumeRequest
		mockLVM     *mockLVM
		expectedErr codes.Code
	}{
		{
			name: "should delete volume successfully",
			req: &csi.DeleteVolumeRequest{
				VolumeId: "test-vg/test-lv",
			},
			mockLVM: &mockLVM{
				deleteLV: func(vg, name string) error {
					return nil
				},
			},
			expectedErr: codes.OK,
		},
		{
			name: "should return success if volume not found",
			req: &csi.DeleteVolumeRequest{
				VolumeId: "test-vg/test-lv",
			},
			mockLVM: &mockLVM{
				deleteLV: func(vg, name string) error {
					return fmt.Errorf("not found")
				},
			},
			expectedErr: codes.OK,
		},
		{
			name: "should fail on invalid volume id",
			req: &csi.DeleteVolumeRequest{
				VolumeId: "invalid-id",
			},
			expectedErr: codes.InvalidArgument,
		},
		{
			name: "should fail on internal error",
			req: &csi.DeleteVolumeRequest{
				VolumeId: "test-vg/test-lv",
			},
			mockLVM: &mockLVM{
				deleteLV: func(vg, name string) error {
					return fmt.Errorf("some other error")
				},
			},
			expectedErr: codes.Internal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			driver := NewDriver("test-endpoint", nil, tt.mockLVM)
			_, err := driver.DeleteVolume(context.Background(), tt.req)
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

func TestControllerExpandVolume(t *testing.T) {
	tests := []struct {
		name        string
		req         *csi.ControllerExpandVolumeRequest
		mockLVM     *mockLVM
		expectedErr codes.Code
	}{
		{
			name: "should expand volume successfully",
			req: &csi.ControllerExpandVolumeRequest{
				VolumeId: "test-vg/test-lv",
				CapacityRange: &csi.CapacityRange{
					RequiredBytes: 2 * 1024 * 1024 * 1024,
				},
			},
			mockLVM: &mockLVM{
				getLV: func(vg, name string) (*lvm.LogicalVolume, error) {
					return &lvm.LogicalVolume{
						Name: "test-lv",
						VG:   "test-vg",
						Size: 1024 * 1024 * 1024,
					}, nil
				},
				resizeLV: func(vg, name string, size int64) error {
					return nil
				},
			},
			expectedErr: codes.OK,
		},
		{
			name: "should return success if volume is already large enough",
			req: &csi.ControllerExpandVolumeRequest{
				VolumeId: "test-vg/test-lv",
				CapacityRange: &csi.CapacityRange{
					RequiredBytes: 1024 * 1024 * 1024,
				},
			},
			mockLVM: &mockLVM{
				getLV: func(vg, name string) (*lvm.LogicalVolume, error) {
					return &lvm.LogicalVolume{
						Name: "test-lv",
						VG:   "test-vg",
						Size: 2 * 1024 * 1024 * 1024,
					}, nil
				},
				resizeLV: func(vg, name string, size int64) error {
					assert.Fail(t, "resizeLV should not have been called")
					return nil
				},
			},
			expectedErr: codes.OK,
		},
		{
			name: "should fail if volume not found",
			req: &csi.ControllerExpandVolumeRequest{
				VolumeId: "test-vg/test-lv",
				CapacityRange: &csi.CapacityRange{
					RequiredBytes: 2 * 1024 * 1024 * 1024,
				},
			},
			mockLVM: &mockLVM{
				getLV: func(vg, name string) (*lvm.LogicalVolume, error) {
					return nil, nil
				},
				resizeLV: func(vg, name string, size int64) error {
					assert.Fail(t, "resizeLV should not have been called")
					return nil
				},
			},
			expectedErr: codes.NotFound,
		},
		{
			name: "should fail on invalid volume id",
			req: &csi.ControllerExpandVolumeRequest{
				VolumeId: "invalid-id",
			},
			expectedErr: codes.InvalidArgument,
		},
		{
			name: "should fail on internal error on get lv",
			req: &csi.ControllerExpandVolumeRequest{
				VolumeId: "test-vg/test-lv",
				CapacityRange: &csi.CapacityRange{
					RequiredBytes: 2 * 1024 * 1024 * 1024,
				},
			},
			mockLVM: &mockLVM{
				getLV: func(vg, name string) (*lvm.LogicalVolume, error) {
					return nil, fmt.Errorf("some error")
				},
				resizeLV: func(vg, name string, size int64) error {
					assert.Fail(t, "resizeLV should not have been called")
					return nil
				},
			},
			expectedErr: codes.Internal,
		},
		{
			name: "should fail on internal error on resize lv",
			req: &csi.ControllerExpandVolumeRequest{
				VolumeId: "test-vg/test-lv",
				CapacityRange: &csi.CapacityRange{
					RequiredBytes: 2 * 1024 * 1024 * 1024,
				},
			},
			mockLVM: &mockLVM{
				getLV: func(vg, name string) (*lvm.LogicalVolume, error) {
					return &lvm.LogicalVolume{
						Name: "test-lv",
						VG:   "test-vg",
						Size: 1024 * 1024 * 1024,
					}, nil
				},
				resizeLV: func(vg, name string, size int64) error {
					return fmt.Errorf("some other error")
				},
			},
			expectedErr: codes.Internal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			driver := NewDriver("test-endpoint", nil, tt.mockLVM)
			_, err := driver.ControllerExpandVolume(context.Background(), tt.req)
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

func TestGetCapacity(t *testing.T) {
	tests := []struct {
		name         string
		req          *csi.GetCapacityRequest
		allowedVGs   []string
		mockLVM      *mockLVM
		expectedErr  codes.Code
		expectedResp *csi.GetCapacityResponse
	}{
		{
			name: "should return capacity for specific VG",
			req: &csi.GetCapacityRequest{
				Parameters: map[string]string{
					volumeGroupKey: "test-vg",
				},
			},
			allowedVGs: []string{"test-vg"},
			mockLVM: &mockLVM{
				getVG: func(name string) (*lvm.VolumeGroup, error) {
					if name == "test-vg" {
						return &lvm.VolumeGroup{
							Name:     "test-vg",
							FreeSize: 100,
						}, nil
					}
					return nil, nil
				},
			},
			expectedErr: codes.OK,
			expectedResp: &csi.GetCapacityResponse{
				AvailableCapacity: 100,
			},
		},
		{
			name:       "should return capacity for allowed VGs (implicit)",
			req:        &csi.GetCapacityRequest{},
			allowedVGs: []string{"vg1", "vg2"},
			mockLVM: &mockLVM{
				getVG: func(name string) (*lvm.VolumeGroup, error) {
					if name == "vg1" {
						return &lvm.VolumeGroup{Name: "vg1", FreeSize: 100}, nil
					}
					if name == "vg2" {
						return &lvm.VolumeGroup{Name: "vg2", FreeSize: 200}, nil
					}
					return nil, nil
				},
			},
			expectedErr: codes.OK,
			expectedResp: &csi.GetCapacityResponse{
				AvailableCapacity: 300,
			},
		},
		{
			name:        "should return 0 if no allowed VGs configured and no param",
			req:         &csi.GetCapacityRequest{},
			allowedVGs:  nil,
			expectedErr: codes.OK,
			expectedResp: &csi.GetCapacityResponse{
				AvailableCapacity: 0,
			},
		},
		{
			name: "should fail if requested VG is not allowed",
			req: &csi.GetCapacityRequest{
				Parameters: map[string]string{
					volumeGroupKey: "forbidden-vg",
				},
			},
			allowedVGs:  []string{"allowed-vg"},
			expectedErr: codes.InvalidArgument,
		},
		{
			name: "should fail if get vg error on specific request",
			req: &csi.GetCapacityRequest{
				Parameters: map[string]string{
					volumeGroupKey: "test-vg",
				},
			},
			allowedVGs: []string{"test-vg"},
			mockLVM: &mockLVM{
				getVG: func(name string) (*lvm.VolumeGroup, error) {
					return nil, fmt.Errorf("lvm error")
				},
			},
			expectedErr: codes.Internal,
		},
		{
			name:       "should skip if get vg error on implicit request",
			req:        &csi.GetCapacityRequest{},
			allowedVGs: []string{"vg1", "error-vg"},
			mockLVM: &mockLVM{
				getVG: func(name string) (*lvm.VolumeGroup, error) {
					if name == "vg1" {
						return &lvm.VolumeGroup{Name: "vg1", FreeSize: 100}, nil
					}
					return nil, fmt.Errorf("lvm error")
				},
			},
			expectedErr: codes.OK,
			expectedResp: &csi.GetCapacityResponse{
				AvailableCapacity: 100,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			driver := NewDriver("test-endpoint", tt.allowedVGs, tt.mockLVM)
			resp, err := driver.GetCapacity(context.Background(), tt.req)
			if tt.expectedErr == codes.OK {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedResp, resp)
			} else {
				assert.Error(t, err)
				st, ok := status.FromError(err)
				assert.True(t, ok)
				assert.Equal(t, tt.expectedErr, st.Code())
			}
		})
	}
}
