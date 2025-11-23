package driver

import (
	"context"
	"fmt"
	"os"

	"github.com/container-storage-interface/spec/lib/go/csi"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"k8s.io/klog/v2"
	"k8s.io/mount-utils"
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

	vgName, lvName, err := getVGAndLVNames(req.VolumeId)
	if err != nil {
		return nil, err
	}

	// check if the volume is already staged
	lv, err := d.lvm.GetLV(vgName, lvName)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get lv: %v", err)
	}
	if lv == nil {
		return nil, status.Errorf(codes.NotFound, "volume '%s' not found", req.VolumeId)
	}

	if !lv.Attr.IsActive() {
		klog.InfoS("Activating LV", "vg", vgName, "lv", lvName)
		if err := d.lvm.ActivateLV(vgName, lvName); err != nil {
			return nil, status.Errorf(codes.Internal, "failed to activate lv: %v", err)
		}
	}

	// skip block volumes
	if req.VolumeCapability.GetBlock() != nil {
		klog.InfoS("Volume is a block device, skipping format and mount", "vg", vgName, "lv", lvName)
		return &csi.NodeStageVolumeResponse{}, nil
	}

	// format and mount the filesystem
	devicePath := fmt.Sprintf("/dev/%s/%s", vgName, lvName)
	fsType := "ext4"
	if mount := req.GetVolumeCapability().GetMount(); mount.GetFsType() != "" {
		fsType = mount.GetFsType()
	}

	if err := d.mounter.FormatAndMount(devicePath, req.StagingTargetPath, fsType, nil); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to format and mount volume: %v", err)
	}

	// check if resize is needed
	needResize, err := d.resizer.NeedResize(devicePath, req.StagingTargetPath)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to check if volume needs resize: %v", err)
	}
	if needResize {
		klog.InfoS("Resizing volume", "devicePath", devicePath, "stagingPath", req.StagingTargetPath)
		if _, err := d.resizer.Resize(devicePath, req.StagingTargetPath); err != nil {
			return nil, status.Errorf(codes.Internal, "failed to resize volume: %v", err)
		}
	}

	return &csi.NodeStageVolumeResponse{}, nil
}

func (d *Driver) NodeUnstageVolume(ctx context.Context, req *csi.NodeUnstageVolumeRequest) (*csi.NodeUnstageVolumeResponse, error) {
	klog.InfoS("NodeUnstageVolume called", "req", req)

	if req.VolumeId == "" {
		return nil, status.Error(codes.InvalidArgument, "volume id is required")
	}
	if req.StagingTargetPath == "" {
		return nil, status.Error(codes.InvalidArgument, "staging target path is required")
	}

	vgName, lvName, err := getVGAndLVNames(req.VolumeId)
	if err != nil {
		return nil, err
	}

	dev, refcnt, err := mount.GetDeviceNameFromMount(d.mounter, req.StagingTargetPath)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to check if staging path is a mount point: %v", err)
	}

	if refcnt == 0 {
		klog.InfoS("Staging path does not exist, assuming unmounted", "stagingPath", req.StagingTargetPath)
		return &csi.NodeUnstageVolumeResponse{}, nil
	}

	if refcnt > 1 {
		klog.InfoS("NodeUnstageVolume: found references to device mounted at target path", "refcnt", refcnt, "device", dev, "stagingPath", req.StagingTargetPath)
	}

	klog.InfoS("Unmounting volume", "stagingPath", req.StagingTargetPath)
	if err := mount.CleanupMountPoint(req.StagingTargetPath, d.mounter, false); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to unmount volume: %v", err)
	}

	// check if the volume was already deactivated
	lv, err := d.lvm.GetLV(vgName, lvName)
	if lv == nil && err == nil {
		// either the vg or the lv are gone - anyways, there's nothing to unstage, so we return success
		klog.InfoS("LV not found, assuming it's already unstaged", "vg", vgName, "lv", lvName)
		return &csi.NodeUnstageVolumeResponse{}, nil

	}
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get lv: %v", err)
	}
	if !lv.Attr.IsActive() {
		klog.InfoS("Volume is already inactive", "vg", vgName, "lv", lvName)
		return &csi.NodeUnstageVolumeResponse{}, nil
	}

	klog.InfoS("Deactivating LV", "vg", vgName, "lv", lvName)
	if err := d.lvm.DeactivateLV(vgName, lvName); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to deactivate lv: %v", err)
	}

	return &csi.NodeUnstageVolumeResponse{}, nil
}

func (d *Driver) NodePublishVolume(ctx context.Context, req *csi.NodePublishVolumeRequest) (*csi.NodePublishVolumeResponse, error) {
	klog.InfoS("NodePublishVolume called", "req", req)

	if req.VolumeId == "" {
		return nil, status.Error(codes.InvalidArgument, "volume id is required")
	}
	if req.StagingTargetPath == "" {
		return nil, status.Error(codes.InvalidArgument, "staging target path is required")
	}
	if req.TargetPath == "" {
		return nil, status.Error(codes.InvalidArgument, "target path is required")
	}
	if req.VolumeCapability == nil {
		return nil, status.Error(codes.InvalidArgument, "volume capability is required")
	}

	// check if the volume is already mounted
	notMnt, err := d.mounter.IsLikelyNotMountPoint(req.TargetPath)
	if err != nil && !os.IsNotExist(err) {
		return nil, status.Errorf(codes.Internal, "failed to check if target path is a mount point: %v", err)
	}
	if !notMnt {
		klog.InfoS("Volume is already mounted", "targetPath", req.TargetPath)
		return &csi.NodePublishVolumeResponse{}, nil
	}

	if req.VolumeCapability.GetBlock() != nil {
		return d.nodePublishVolumeBlock(req)
	}
	return d.nodePublishVolumeMount(req)
}

func (d *Driver) nodePublishVolumeMount(req *csi.NodePublishVolumeRequest) (*csi.NodePublishVolumeResponse, error) {
	klog.InfoS("Publishing mount volume", "volumeId", req.VolumeId, "targetPath", req.TargetPath)

	// ensure target path exists
	if err := os.MkdirAll(req.TargetPath, 0755); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create target path: %v", err)
	}

	options := []string{"bind"}
	if req.Readonly {
		options = append(options, "ro")
	}

	fsType := req.GetVolumeCapability().GetMount().GetFsType()

	if err := d.mounter.Mount(req.StagingTargetPath, req.TargetPath, fsType, options); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to mount volume: %v", err)
	}

	return &csi.NodePublishVolumeResponse{}, nil
}

func (d *Driver) nodePublishVolumeBlock(req *csi.NodePublishVolumeRequest) (*csi.NodePublishVolumeResponse, error) {
	klog.InfoS("Publishing block volume", "volumeId", req.VolumeId, "targetPath", req.TargetPath)

	devicePath, err := getDevicePath(req.VolumeId)
	if err != nil {
		return nil, err
	}

	// ensure target path exists
	if _, err := os.Stat(req.TargetPath); os.IsNotExist(err) {
		// create the file
		f, err := os.OpenFile(req.TargetPath, os.O_CREATE|os.O_RDWR, 0644)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "failed to create target path file: %v", err)
		}
		if err := f.Close(); err != nil {
			return nil, status.Errorf(codes.Internal, "failed to close target path file: %v", err)
		}
	}

	options := []string{"bind"}
	if req.Readonly {
		options = append(options, "ro")
	}

	if err := d.mounter.Mount(devicePath, req.TargetPath, "", options); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to mount volume: %v", err)
	}

	return &csi.NodePublishVolumeResponse{}, nil
}

func (d *Driver) NodeUnpublishVolume(ctx context.Context, req *csi.NodeUnpublishVolumeRequest) (*csi.NodeUnpublishVolumeResponse, error) {
	klog.InfoS("NodeUnpublishVolume called", "req", req)

	if req.VolumeId == "" {
		return nil, status.Error(codes.InvalidArgument, "volume id is required")
	}
	if req.TargetPath == "" {
		return nil, status.Error(codes.InvalidArgument, "target path is required")
	}

	// Check if the target path is a mount point
	notMnt, err := d.mounter.IsLikelyNotMountPoint(req.TargetPath)
	if err != nil {
		if os.IsNotExist(err) {
			klog.InfoS("Target path does not exist, assuming unmounted", "targetPath", req.TargetPath)
			return &csi.NodeUnpublishVolumeResponse{}, nil
		}
		return nil, status.Errorf(codes.Internal, "failed to check if target path is a mount point: %v", err)
	}
	if notMnt {
		klog.InfoS("Target path is not a mount point, assuming unmounted", "targetPath", req.TargetPath)
		return &csi.NodeUnpublishVolumeResponse{}, nil
	}

	klog.InfoS("Unmounting volume", "targetPath", req.TargetPath)
	if err := d.mounter.Unmount(req.TargetPath); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to unmount volume: %v", err)
	}

	// Remove the target path
	if err := os.Remove(req.TargetPath); err != nil && !os.IsNotExist(err) {
		return nil, status.Errorf(codes.Internal, "failed to remove target path: %v", err)
	}

	return &csi.NodeUnpublishVolumeResponse{}, nil
}

func (d *Driver) NodeGetVolumeStats(ctx context.Context, req *csi.NodeGetVolumeStatsRequest) (*csi.NodeGetVolumeStatsResponse, error) {
	klog.InfoS("NodeGetVolumeStats called", "req", req)

	if req.VolumeId == "" {
		return nil, status.Error(codes.InvalidArgument, "volume id is required")
	}
	if req.VolumePath == "" {
		return nil, status.Error(codes.InvalidArgument, "volume path is required")
	}

	if _, err := os.Stat(req.VolumePath); os.IsNotExist(err) {
		return nil, status.Errorf(codes.NotFound, "volume path %s not found", req.VolumePath)
	}

	isBlock, err := d.stats.IsBlockDevice(req.VolumePath)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to check if path is block device: %v", err)
	}

	if isBlock {
		totalBytes, err := d.stats.GetBlockSizeBytes(req.VolumePath)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "failed to get block device stats: %v", err)
		}
		return &csi.NodeGetVolumeStatsResponse{
			Usage: []*csi.VolumeUsage{
				{
					Total: totalBytes,
					Unit:  csi.VolumeUsage_BYTES,
				},
			},
		}, nil
	}

	available, capacity, used, inodes, inodesFree, inodesUsed, err := d.stats.GetFSStats(req.VolumePath)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get fs stats: %v", err)
	}

	return &csi.NodeGetVolumeStatsResponse{
		Usage: []*csi.VolumeUsage{
			{
				Available: available,
				Total:     capacity,
				Used:      used,
				Unit:      csi.VolumeUsage_BYTES,
			},
			{
				Available: inodesFree,
				Total:     inodes,
				Used:      inodesUsed,
				Unit:      csi.VolumeUsage_INODES,
			},
		},
	}, nil
}

func (d *Driver) NodeExpandVolume(ctx context.Context, req *csi.NodeExpandVolumeRequest) (*csi.NodeExpandVolumeResponse, error) {
	klog.InfoS("NodeExpandVolume called", "req", req)

	if req.VolumeId == "" {
		return nil, status.Error(codes.InvalidArgument, "volume id is required")
	}
	if req.VolumePath == "" {
		return nil, status.Error(codes.InvalidArgument, "volume path is required")
	}

	devicePath, err := getDevicePath(req.VolumeId)
	if err != nil {
		return nil, err
	}

	// skip block volumes
	if req.VolumeCapability.GetBlock() != nil {
		klog.InfoS("Volume is a block device, skipping filesystem resize")
		return &csi.NodeExpandVolumeResponse{}, nil
	}

	klog.InfoS("Resizing filesystem", "devicePath", devicePath, "volumePath", req.VolumePath)
	if _, err := d.resizer.Resize(devicePath, req.VolumePath); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to resize filesystem: %v", err)
	}

	return &csi.NodeExpandVolumeResponse{}, nil
}

func (d *Driver) NodeGetCapabilities(ctx context.Context, req *csi.NodeGetCapabilitiesRequest) (*csi.NodeGetCapabilitiesResponse, error) {
	klog.InfoS("NodeGetCapabilities called", "req", req)
	return &csi.NodeGetCapabilitiesResponse{
		Capabilities: []*csi.NodeServiceCapability{
			{
				Type: &csi.NodeServiceCapability_Rpc{
					Rpc: &csi.NodeServiceCapability_RPC{
						Type: csi.NodeServiceCapability_RPC_STAGE_UNSTAGE_VOLUME,
					},
				},
			},
			{
				Type: &csi.NodeServiceCapability_Rpc{
					Rpc: &csi.NodeServiceCapability_RPC{
						Type: csi.NodeServiceCapability_RPC_EXPAND_VOLUME,
					},
				},
			},
			{
				Type: &csi.NodeServiceCapability_Rpc{
					Rpc: &csi.NodeServiceCapability_RPC{
						Type: csi.NodeServiceCapability_RPC_GET_VOLUME_STATS,
					},
				},
			},
		},
	}, nil
}

func (d *Driver) NodeGetInfo(ctx context.Context, req *csi.NodeGetInfoRequest) (*csi.NodeGetInfoResponse, error) {
	klog.InfoS("NodeGetInfo called", "req", req)
	return nil, status.Error(codes.Unimplemented, "")
}
