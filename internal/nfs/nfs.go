package nfs

import (
	"os"

	"github.com/chenliu1993/simple-csi-driver/internal/server"
	"github.com/container-storage-interface/spec/lib/go/csi"
	"k8s.io/klog/v2"
)

var (
	controllerCapsList = []csi.ControllerServiceCapability_RPC_Type{
		csi.ControllerServiceCapability_RPC_CREATE_DELETE_VOLUME,
		csi.ControllerServiceCapability_RPC_PUBLISH_UNPUBLISH_VOLUME,
		csi.ControllerServiceCapability_RPC_SINGLE_NODE_MULTI_WRITER,
		// csi.ControllerServiceCapability_RPC_EXPAND_VOLUME,
	}

	nodeCapsList = []csi.NodeServiceCapability_RPC_Type{
		csi.NodeServiceCapability_RPC_GET_VOLUME_STATS,
		csi.NodeServiceCapability_RPC_SINGLE_NODE_MULTI_WRITER,
		csi.NodeServiceCapability_RPC_UNKNOWN,
	}
)

type nfsDriver struct {
	name     string
	endpoint string
	node     string

	ids csi.IdentityServer
	cs  csi.ControllerServer
	ns  csi.NodeServer

	controllerCaps []*csi.ControllerServiceCapability
	nodeCaps       []*csi.NodeServiceCapability

	stopCh chan os.Signal
}

func NewNFSDriver(driverName, endpoint, node string, stopCh chan os.Signal) *nfsDriver {
	klog.V(4).InfoS("Starting nfs driver...")
	nfsClient := &nfsDriver{
		name:     driverName,
		endpoint: endpoint,
		node:     node,
		stopCh:   stopCh,
	}

	nfsClient.ids = NewIdentityServer(nfsClient)
	nfsClient.cs = NewControllerServer(nfsClient)
	nfsClient.ns = NewNodeServer(nfsClient)

	nfsClient.AddControllerCapabilities(controllerCapsList)
	nfsClient.AddNodeCapabilities(nodeCapsList)

	return nfsClient
}

func (nd *nfsDriver) Run() {
	s := server.NewNonBlockingGRPCServer()
	s.Start(nd.endpoint,
		nd.ids,
		nd.cs,
		nd.ns,
	)
	go func() {
		<-nd.stopCh
		klog.V(4).InfoS("Stopping nfs driver...")
		s.Stop()
		klog.Flush()
	}()

	s.Wait()
}

func (nd *nfsDriver) AddControllerCapabilities(caps []csi.ControllerServiceCapability_RPC_Type) {
	var ndCtrCaps []*csi.ControllerServiceCapability
	for _, cap := range caps {
		ndCtrCaps = append(ndCtrCaps, NewControllerServiceCapability(cap))
	}
	nd.controllerCaps = append(nd.controllerCaps, ndCtrCaps...)
}

func (nd *nfsDriver) AddNodeCapabilities(caps []csi.NodeServiceCapability_RPC_Type) {
	var ndNodeCaps []*csi.NodeServiceCapability
	for _, cap := range caps {
		ndNodeCaps = append(ndNodeCaps, NewNodeServiceCapability(cap))
	}
	nd.nodeCaps = append(nd.nodeCaps, ndNodeCaps...)
}

func NewControllerServiceCapability(cap csi.ControllerServiceCapability_RPC_Type) *csi.ControllerServiceCapability {
	return &csi.ControllerServiceCapability{
		Type: &csi.ControllerServiceCapability_Rpc{
			Rpc: &csi.ControllerServiceCapability_RPC{
				Type: cap,
			},
		},
	}
}

func NewNodeServiceCapability(cap csi.NodeServiceCapability_RPC_Type) *csi.NodeServiceCapability {
	return &csi.NodeServiceCapability{
		Type: &csi.NodeServiceCapability_Rpc{
			Rpc: &csi.NodeServiceCapability_RPC{
				Type: cap,
			},
		},
	}
}
