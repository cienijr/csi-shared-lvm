package driver

import (
	"github.com/container-storage-interface/spec/lib/go/csi"
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
}

func NewDriver(endpoint string, allowedVolumeGroups []string, lvm lvm.LVM) *Driver {
	return &Driver{
		endpoint:            endpoint,
		allowedVolumeGroups: allowedVolumeGroups,
		lvm:                 lvm,
		mounter:             &mount.SafeFormatAndMount{Interface: mount.New(""), Exec: utilexec.New()},
	}
}
