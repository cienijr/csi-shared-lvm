package driver

import "github.com/container-storage-interface/spec/lib/go/csi"

type Driver struct {
	csi.UnimplementedIdentityServer
	csi.UnimplementedControllerServer
	csi.UnimplementedNodeServer

	endpoint            string
	allowedVolumeGroups []string
}

func NewDriver(endpoint string, allowedVolumeGroups []string) *Driver {
	return &Driver{
		endpoint:            endpoint,
		allowedVolumeGroups: allowedVolumeGroups,
	}
}
