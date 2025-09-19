package driver

import (
	"context"

	"github.com/container-storage-interface/spec/lib/go/csi"
	"k8s.io/klog/v2"
)

func (d *Driver) GetPluginInfo(ctx context.Context, req *csi.GetPluginInfoRequest) (*csi.GetPluginInfoResponse, error) {
	klog.InfoS("GetPluginInfo called", "req", req)
	return &csi.GetPluginInfoResponse{
		Name:          "csi-shared-lvm.cienijr.github.com",
		VendorVersion: "0.1.0",
	}, nil
}

func (d *Driver) GetPluginCapabilities(ctx context.Context, req *csi.GetPluginCapabilitiesRequest) (*csi.GetPluginCapabilitiesResponse, error) {
	klog.InfoS("GetPluginCapabilities called", "req", req)
	return &csi.GetPluginCapabilitiesResponse{
		Capabilities: []*csi.PluginCapability{
			{
				Type: &csi.PluginCapability_Service_{
					Service: &csi.PluginCapability_Service{
						Type: csi.PluginCapability_Service_CONTROLLER_SERVICE,
					},
				},
			},
		},
	}, nil
}

func (d *Driver) Probe(ctx context.Context, req *csi.ProbeRequest) (*csi.ProbeResponse, error) {
	klog.InfoS("Probe called", "req", req)
	return &csi.ProbeResponse{}, nil
}
