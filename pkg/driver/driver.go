package driver

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/container-storage-interface/spec/lib/go/csi"
	"golang.org/x/sys/unix"
	"k8s.io/mount-utils"
	utilexec "k8s.io/utils/exec"

	"github.com/cienijr/csi-shared-lvm/pkg/lvm"
)

type Driver struct {
	csi.UnimplementedIdentityServer
	csi.UnimplementedControllerServer
	csi.UnimplementedNodeServer

	endpoint            string
	allowedVolumeGroups []string
	lvm                 lvm.LVM
	mounter             *mount.SafeFormatAndMount
	resizer             Resizer
	stats               DeviceStats
}

type Resizer interface {
	NeedResize(devicePath, deviceMountPath string) (bool, error)
	Resize(devicePath, deviceMountPath string) (bool, error)
}

type DeviceStats interface {
	GetBlockSizeBytes(devicePath string) (int64, error)
	GetFSStats(path string) (available, capacity, used, inodes, inodesFree, inodesUsed int64, err error)
	IsBlockDevice(path string) (bool, error)
}

type defaultDeviceStats struct {
	exec utilexec.Interface
}

func (d *defaultDeviceStats) GetBlockSizeBytes(devicePath string) (int64, error) {
	output, err := d.exec.Command("blockdev", "--getsize64", devicePath).CombinedOutput()
	if err != nil {
		return 0, fmt.Errorf("failed to get block size: %v, output: %s", err, string(output))
	}
	strOut := strings.TrimSpace(string(output))
	return strconv.ParseInt(strOut, 10, 64)
}

func (d *defaultDeviceStats) GetFSStats(path string) (int64, int64, int64, int64, int64, int64, error) {
	var statfs unix.Statfs_t
	if err := unix.Statfs(path, &statfs); err != nil {
		return 0, 0, 0, 0, 0, 0, err
	}

	available := int64(statfs.Bavail) * int64(statfs.Bsize)
	capacity := int64(statfs.Blocks) * int64(statfs.Bsize)
	used := (int64(statfs.Blocks) - int64(statfs.Bfree)) * int64(statfs.Bsize)

	inodes := int64(statfs.Files)
	inodesFree := int64(statfs.Ffree)
	inodesUsed := inodes - inodesFree

	return available, capacity, used, inodes, inodesFree, inodesUsed, nil
}

func (d *defaultDeviceStats) IsBlockDevice(path string) (bool, error) {
	info, err := os.Stat(path)
	if err != nil {
		return false, err
	}
	return (info.Mode() & os.ModeDevice) != 0, nil
}

func NewDriver(endpoint string, allowedVolumeGroups []string, lvm lvm.LVM) *Driver {
	mountExec := utilexec.New()
	return &Driver{
		endpoint:            endpoint,
		allowedVolumeGroups: allowedVolumeGroups,
		lvm:                 lvm,
		mounter:             &mount.SafeFormatAndMount{Interface: mount.New(""), Exec: mountExec},
		resizer:             mount.NewResizeFs(mountExec),
		stats:               &defaultDeviceStats{exec: mountExec},
	}
}
