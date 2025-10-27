package driver

import (
	"context"
	"strings"

	"github.com/container-storage-interface/spec/lib/go/csi"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"k8s.io/klog/v2"
)

func (d *Driver) NodeStageVolume(ctx context.Context, req *csi.NodeStageVolumeRequest) (*csi.NodeStageVolumeResponse, error) {
	klog.InfoS("NodeStageVolume called", "req", req)

	if req.VolumeId == "" {
		return nil, status.Error(codes.InvalidArgument, "volume id is required")
	}
	if req.StagingTargetPath == "" {
		return nil, status.Error(codes.InvalidArgument, "staging target path is required")
	}
	if req.VolumeCapability == nil {
		return nil, status.Error(codes.InvalidArgument, "volume capability is required")
	}

	parts := strings.Split(req.VolumeId, "/")
	if len(parts) != 2 {
		return nil, status.Errorf(codes.InvalidArgument, "invalid volume id: %s", req.VolumeId)
	}
	vgName, lvName := parts[0], parts[1]

	// check if the volume is already staged
	lv, err := d.lvm.GetLV(vgName, lvName)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get lv: %v", err)
	}
	if lv == nil {
		return nil, status.Errorf(codes.NotFound, "volume '%s' not found", req.VolumeId)
	}

	if lv.Attr.IsActive() {
		klog.InfoS("Volume is already active", "vg", vgName, "lv", lvName)
		return &csi.NodeStageVolumeResponse{}, nil
	}

	klog.InfoS("Activating LV", "vg", vgName, "lv", lvName)
	if err := d.lvm.ActivateLV(vgName, lvName); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to activate lv: %v", err)
	}

	return &csi.NodeStageVolumeResponse{}, nil
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
