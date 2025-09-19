package driver

import "github.com/container-storage-interface/spec/lib/go/csi"

type Driver struct {
	csi.UnimplementedIdentityServer
	csi.UnimplementedControllerServer
	csi.UnimplementedNodeServer

	endpoint string
}

func NewDriver(endpoint string) *Driver {
	return &Driver{
		endpoint: endpoint,
	}
}
