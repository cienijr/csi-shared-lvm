package driver

import (
	"github.com/container-storage-interface/spec/lib/go/csi"

	"github.com/cienijr/csi-shared-lvm/pkg/lvm"
)

type Driver struct {
	csi.UnimplementedIdentityServer
	csi.UnimplementedControllerServer
	csi.UnimplementedNodeServer

	endpoint            string
	allowedVolumeGroups []string
	lvm                 lvm.LVM
}

func NewDriver(endpoint string, allowedVolumeGroups []string, lvm lvm.LVM) *Driver {
	return &Driver{
		endpoint:            endpoint,
		allowedVolumeGroups: allowedVolumeGroups,
		lvm:                 lvm,
	}
}
