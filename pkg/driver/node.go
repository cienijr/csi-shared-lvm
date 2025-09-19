package driver

import (
	"context"

	"github.com/container-storage-interface/spec/lib/go/csi"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"k8s.io/klog/v2"
)

func (d *Driver) NodeStageVolume(ctx context.Context, req *csi.NodeStageVolumeRequest) (*csi.NodeStageVolumeResponse, error) {
	klog.InfoS("NodeStageVolume called", "req", req)
	return nil, status.Error(codes.Unimplemented, "")
}

func (d *Driver) NodeUnstageVolume(ctx context.Context, req *csi.NodeUnstageVolumeRequest) (*csi.NodeUnstageVolumeResponse, error) {
	klog.InfoS("NodeUnstageVolume called", "req", req)
	return nil, status.Error(codes.Unimplemented, "")
}

func (d *Driver) NodePublishVolume(ctx context.Context, req *csi.NodePublishVolumeRequest) (*csi.NodePublishVolumeResponse, error) {
	klog.InfoS("NodePublishVolume called", "req", req)
	return nil, status.Error(codes.Unimplemented, "")
}

func (d *Driver) NodeUnpublishVolume(ctx context.Context, req *csi.NodeUnpublishVolumeRequest) (*csi.NodeUnpublishVolumeResponse, error) {
	klog.InfoS("NodeUnpublishVolume called", "req", req)
	return nil, status.Error(codes.Unimplemented, "")
}

func (d *Driver) NodeGetVolumeStats(ctx context.Context, req *csi.NodeGetVolumeStatsRequest) (*csi.NodeGetVolumeStatsResponse, error) {
	klog.InfoS("NodeGetVolumeStats called", "req", req)
	return nil, status.Error(codes.Unimplemented, "")
}

func (d *Driver) NodeExpandVolume(ctx context.Context, req *csi.NodeExpandVolumeRequest) (*csi.NodeExpandVolumeResponse, error) {
	klog.InfoS("NodeExpandVolume called", "req", req)
	return nil, status.Error(codes.Unimplemented, "")
}

func (d *Driver) NodeGetCapabilities(ctx context.Context, req *csi.NodeGetCapabilitiesRequest) (*csi.NodeGetCapabilitiesResponse, error) {
	klog.InfoS("NodeGetCapabilities called", "req", req)
	return nil, status.Error(codes.Unimplemented, "")
}

func (d *Driver) NodeGetInfo(ctx context.Context, req *csi.NodeGetInfoRequest) (*csi.NodeGetInfoResponse, error) {
	klog.InfoS("NodeGetInfo called", "req", req)
	return nil, status.Error(codes.Unimplemented, "")
}
