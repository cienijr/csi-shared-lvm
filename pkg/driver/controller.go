package driver

import (
	"context"

	"github.com/container-storage-interface/spec/lib/go/csi"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"k8s.io/klog/v2"
)

func (d *Driver) CreateVolume(ctx context.Context, req *csi.CreateVolumeRequest) (*csi.CreateVolumeResponse, error) {
	klog.InfoS("CreateVolume called", "req", req)
	return nil, status.Error(codes.Unimplemented, "")
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
