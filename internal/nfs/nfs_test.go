package nfs

import (
	"fmt"
	"testing"

	csi "github.com/container-storage-interface/spec/lib/go/csi"
	"github.com/stretchr/testify/assert"
)

const (
	fakeNode     = "fakeNode"
	fakeEndpoint = "fakeEndpoint"
)

func NewFakeNfsDriver(node string) *nfsDriver {
	return &nfsDriver{
		endpoint: fakeEndpoint,
		node:     node,
	}
}

func TestNewFakeNfsDriver(t *testing.T) {
	d := NewFakeNfsDriver(fakeNode)
	assert.Equal(t, fakeEndpoint, d.endpoint)
	assert.Equal(t, fakeNode, d.node)
}

func TestNewControllerServiceCapability(t *testing.T) {
	tcs := []struct {
		description  string
		capability   csi.ControllerServiceCapability_RPC_Type
		expectedCaps *csi.ControllerServiceCapability
	}{
		{
			description: "CREATE_DELETE_VOLUME",
			capability:  csi.ControllerServiceCapability_RPC_CREATE_DELETE_VOLUME,
			expectedCaps: &csi.ControllerServiceCapability{
				Type: &csi.ControllerServiceCapability_Rpc{
					Rpc: &csi.ControllerServiceCapability_RPC{
						Type: csi.ControllerServiceCapability_RPC_CREATE_DELETE_VOLUME,
					},
				},
			},
		},
		{
			description: "PUBLISH_UNPUBLISH_VOLUME",
			capability:  csi.ControllerServiceCapability_RPC_PUBLISH_UNPUBLISH_VOLUME,
			expectedCaps: &csi.ControllerServiceCapability{
				Type: &csi.ControllerServiceCapability_Rpc{
					Rpc: &csi.ControllerServiceCapability_RPC{
						Type: csi.ControllerServiceCapability_RPC_PUBLISH_UNPUBLISH_VOLUME,
					},
				},
			},
		},
		{

			description: "EXPAND_VOLUME",
			capability:  csi.ControllerServiceCapability_RPC_EXPAND_VOLUME,
			expectedCaps: &csi.ControllerServiceCapability{
				Type: &csi.ControllerServiceCapability_Rpc{
					Rpc: &csi.ControllerServiceCapability_RPC{
						Type: csi.ControllerServiceCapability_RPC_EXPAND_VOLUME,
					},
				},
			},
		},
		{

			description: "SINGLE_NODE_MULTI_WRITER",
			capability:  csi.ControllerServiceCapability_RPC_SINGLE_NODE_MULTI_WRITER,
			expectedCaps: &csi.ControllerServiceCapability{
				Type: &csi.ControllerServiceCapability_Rpc{
					Rpc: &csi.ControllerServiceCapability_RPC{
						Type: csi.ControllerServiceCapability_RPC_SINGLE_NODE_MULTI_WRITER,
					},
				},
			},
		},
	}

	for _, tc := range tcs {
		t.Run(tc.description, func(t *testing.T) {
			actualCaps := NewControllerServiceCapability(tc.capability)
			if actualCaps == nil {
				t.Fatalf("NewControllerServiceCapability returned nil")
			}
			assert.Equal(t, tc.expectedCaps, actualCaps, fmt.Sprintf("Expected: %v, got: %v", tc.expectedCaps, actualCaps))
		})
	}
}

func TestNewNodeServiceCapability(t *testing.T) {
	tcs := []struct {
		description  string
		capability   csi.NodeServiceCapability_RPC_Type
		expectedCaps *csi.NodeServiceCapability
	}{
		{
			description: "SINGLE_NODE_MULTI_WRITER",
			capability:  csi.NodeServiceCapability_RPC_SINGLE_NODE_MULTI_WRITER,
			expectedCaps: &csi.NodeServiceCapability{
				Type: &csi.NodeServiceCapability_Rpc{
					Rpc: &csi.NodeServiceCapability_RPC{
						Type: csi.NodeServiceCapability_RPC_SINGLE_NODE_MULTI_WRITER,
					},
				},
			},
		},
		{
			description: "GET_VOLUME_STATS",
			capability:  csi.NodeServiceCapability_RPC_GET_VOLUME_STATS,
			expectedCaps: &csi.NodeServiceCapability{
				Type: &csi.NodeServiceCapability_Rpc{
					Rpc: &csi.NodeServiceCapability_RPC{
						Type: csi.NodeServiceCapability_RPC_GET_VOLUME_STATS,
					},
				},
			},
		},
		{
			description: "UNKNOWN",
			capability:  csi.NodeServiceCapability_RPC_UNKNOWN,
			expectedCaps: &csi.NodeServiceCapability{
				Type: &csi.NodeServiceCapability_Rpc{
					Rpc: &csi.NodeServiceCapability_RPC{
						Type: csi.NodeServiceCapability_RPC_UNKNOWN,
					},
				},
			},
		},
	}

	for _, tc := range tcs {
		t.Run(tc.description, func(t *testing.T) {
			actualCaps := NewNodeServiceCapability(tc.capability)
			if actualCaps == nil {
				t.Fatalf("NewNodeServiceCapability returned nil")
			}
			assert.Equal(t, tc.expectedCaps, actualCaps, fmt.Sprintf("Expected: %v, got: %v", tc.expectedCaps, actualCaps))
		})
	}
}
