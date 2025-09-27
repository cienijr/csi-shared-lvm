package driver

import (
	"context"
	"fmt"

	"github.com/container-storage-interface/spec/lib/go/csi"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"k8s.io/klog/v2"

	"github.com/cienijr/csi-shared-lvm/pkg/lvm"
)

const (
	volumeGroupKey = "volumeGroup"
)

func (d *Driver) CreateVolume(ctx context.Context, req *csi.CreateVolumeRequest) (*csi.CreateVolumeResponse, error) {
	klog.InfoS("CreateVolume called", "req", req)

	lvName := req.Name
	if lvName == "" {
		return nil, status.Error(codes.InvalidArgument, "name is required")
	}
	if req.VolumeCapabilities == nil || len(req.VolumeCapabilities) == 0 {
		return nil, status.Error(codes.InvalidArgument, "volume capabilities are required")
	}
	if req.CapacityRange == nil {
		return nil, status.Error(codes.InvalidArgument, "capacity range is required")
	}

	params := req.GetParameters()
	vgName, ok := params[volumeGroupKey]
	if !ok {
		return nil, status.Errorf(codes.InvalidArgument, "parameter '%s' is required", volumeGroupKey)
	}

	if len(d.allowedVolumeGroups) > 0 {
		allowed := false
		for _, vg := range d.allowedVolumeGroups {
			if vg == vgName {
				allowed = true
				break
			}
		}
		if !allowed {
			return nil, status.Errorf(codes.InvalidArgument, "volume group '%s' is not allowed", vgName)
		}
	}

	size := req.GetCapacityRange().GetRequiredBytes()

	lv, err := lvm.GetLV(vgName, lvName)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get lv: %v", err)
	}

	if lv != nil {
		// idempotency
		if lv.Size >= size {
			klog.InfoS("LV already exists and is large enough, returning success", "vg", vgName, "lv", lvName)
			return &csi.CreateVolumeResponse{
				Volume: &csi.Volume{
					VolumeId:      fmt.Sprintf("%s/%s", vgName, lvName),
					CapacityBytes: lv.Size,
				},
			}, nil
		}
		return nil, status.Errorf(codes.AlreadyExists, "lv '%s' already exists but with a different size", lvName)
	}

	klog.InfoS("Creating new LV", "vg", vgName, "lv", lvName, "size", size)
	tags := []string{
		lvm.OwnershipTag,
	}
	if err := lvm.CreateLV(vgName, lvName, size, tags); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create lv: %v", err)
	}

	// actual volume size may be higher than requested, since LVM rounds up to 4MiB sectors
	actualLV, err := lvm.GetLV(vgName, lvName)
	if err != nil || actualLV == nil {
		return nil, status.Errorf(codes.Internal, "failed to get lv after creation: %v", err)
	}

	actualSize := actualLV.Size

	return &csi.CreateVolumeResponse{
		Volume: &csi.Volume{
			VolumeId:      fmt.Sprintf("%s/%s", vgName, lvName),
			CapacityBytes: actualSize,
		},
	}, nil
}

func (d *Driver) DeleteVolume(ctx context.Context, req *csi.DeleteVolumeRequest) (*csi.DeleteVolumeResponse, error) {
	klog.InfoS("DeleteVolume called", "req", req)
	return nil, status.Error(codes.Unimplemented, "")
}

func (d *Driver) ControllerPublishVolume(ctx context.Context, req *csi.ControllerPublishVolumeRequest) (*csi.ControllerPublishVolumeResponse, error) {
	klog.InfoS("ControllerPublishVolume called", "req", req)
	return nil, status.Error(codes.Unimplemented, "")
}

func (d *Driver) ControllerUnpublishVolume(ctx context.Context, req *csi.ControllerUnpublishVolumeRequest) (*csi.ControllerUnpublishVolumeResponse, error) {
	klog.InfoS("ControllerUnpublishVolume called", "req", req)
	return nil, status.Error(codes.Unimplemented, "")
}

func (d *Driver) ValidateVolumeCapabilities(ctx context.Context, req *csi.ValidateVolumeCapabilitiesRequest) (*csi.ValidateVolumeCapabilitiesResponse, error) {
	klog.InfoS("ValidateVolumeCapabilities called", "req", req)
	return nil, status.Error(codes.Unimplemented, "")
}

func (d *Driver) ListVolumes(ctx context.Context, req *csi.ListVolumesRequest) (*csi.ListVolumesResponse, error) {
	klog.InfoS("ListVolumes called", "req", req)
	return nil, status.Error(codes.Unimplemented, "")
}

func (d *Driver) GetCapacity(ctx context.Context, req *csi.GetCapacityRequest) (*csi.GetCapacityResponse, error) {
	klog.InfoS("GetCapacity called", "req", req)
	return nil, status.Error(codes.Unimplemented, "")
}

func (d *Driver) ControllerGetCapabilities(ctx context.Context, req *csi.ControllerGetCapabilitiesRequest) (*csi.ControllerGetCapabilitiesResponse, error) {
	klog.InfoS("ControllerGetCapabilities called", "req", req)
	return nil, status.Error(codes.Unimplemented, "")
}

func (d *Driver) CreateSnapshot(ctx context.Context, req *csi.CreateSnapshotRequest) (*csi.CreateSnapshotResponse, error) {
	klog.InfoS("CreateSnapshot called", "req", req)
	return nil, status.Error(codes.Unimplemented, "")
}

func (d *Driver) DeleteSnapshot(ctx context.Context, req *csi.DeleteSnapshotRequest) (*csi.DeleteSnapshotResponse, error) {
	klog.InfoS("DeleteSnapshot called", "req", req)
	return nil, status.Error(codes.Unimplemented, "")
}

func (d *Driver) ListSnapshots(ctx context.Context, req *csi.ListSnapshotsRequest) (*csi.ListSnapshotsResponse, error) {
	klog.InfoS("ListSnapshots called", "req", req)
	return nil, status.Error(codes.Unimplemented, "")
}

func (d *Driver) ControllerExpandVolume(ctx context.Context, req *csi.ControllerExpandVolumeRequest) (*csi.ControllerExpandVolumeResponse, error) {
	klog.InfoS("ControllerExpandVolume called", "req", req)
	return nil, status.Error(codes.Unimplemented, "")
}

func (d *Driver) ControllerGetVolume(ctx context.Context, req *csi.ControllerGetVolumeRequest) (*csi.ControllerGetVolumeResponse, error) {
	klog.InfoS("ControllerGetVolume called", "req", req)
	return nil, status.Error(codes.Unimplemented, "")
}

func (d *Driver) ControllerModifyVolume(ctx context.Context, req *csi.ControllerModifyVolumeRequest) (*csi.ControllerModifyVolumeResponse, error) {
	klog.InfoS("ControllerModifyVolume called", "req", req)
	return nil, status.Error(codes.Unimplemented, "")
}
