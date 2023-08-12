package nfs

import (
	"context"

	csi "github.com/container-storage-interface/spec/lib/go/csi"
	"k8s.io/klog/v2"
)

var (
	// This will get decided while build using commit or tag
	nfsDriverVersion = "0.0.0"
)

// Check if implements csi.IdentityServer
var _ csi.IdentityServer = &identityServer{}

type identityServer struct {
	// Embed the default IdentityServer
	driver *nfsDriver
}

// GetPluginInfo implements csi.IdentityServer.
func (i *identityServer) GetPluginInfo(context.Context, *csi.GetPluginInfoRequest) (*csi.GetPluginInfoResponse, error) {
	klog.V(4).InfoS("Getting nfsdriver info......")
	return &csi.GetPluginInfoResponse{
		Name: i.driver.name,
		// TODO: currently let me fake one version, will use driver version later
		VendorVersion: nfsDriverVersion,
	}, nil
}

func NewIdentityServer(driver *nfsDriver) *identityServer {
	return &identityServer{
		driver: driver,
	}
}

// GetPluginCapabilities returens the capabilities of the nfs driver.
func (i *identityServer) GetPluginCapabilities(ctx context.Context, req *csi.GetPluginCapabilitiesRequest) (*csi.GetPluginCapabilitiesResponse, error) {
	klog.V(4).InfoS("Getting nfsdriver capacities......")
	return &csi.GetPluginCapabilitiesResponse{
		Capabilities: []*csi.PluginCapability{
			{
				Type: &csi.PluginCapability_Service_{
					Service: &csi.PluginCapability_Service{
						Type: csi.PluginCapability_Service_CONTROLLER_SERVICE,
					},
				},
			},
			{
				Type: &csi.PluginCapability_Service_{
					Service: &csi.PluginCapability_Service{
						Type: csi.PluginCapability_Service_VOLUME_ACCESSIBILITY_CONSTRAINTS,
					},
				},
			},
			{
				Type: &csi.PluginCapability_VolumeExpansion_{
					VolumeExpansion: &csi.PluginCapability_VolumeExpansion{
						Type: csi.PluginCapability_VolumeExpansion_ONLINE,
					},
				},
			},
		},
	}, nil
}

// Probe is used to check the health status of nfs driver
func (i *identityServer) Probe(ctx context.Context, req *csi.ProbeRequest) (*csi.ProbeResponse, error) {
	klog.V(4).InfoS("Probing nfsdriver health status......")
	return &csi.ProbeResponse{}, nil
}
