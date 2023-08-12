package nfs

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	csi "github.com/container-storage-interface/spec/lib/go/csi"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"k8s.io/klog/v2"
	"k8s.io/kubernetes/pkg/volume"
	mount "k8s.io/mount-utils"
)

// Check if implements csi.NodeServer
var _ csi.NodeServer = &nodeServer{}

type nodeServer struct {
	driver *nfsDriver

	mounter mount.Interface
}

// NewNodeServer returens a functional node server
func NewNodeServer(driver *nfsDriver) *nodeServer {
	return &nodeServer{
		driver:  driver,
		mounter: mount.New(""),
	}
}

func (ns *nodeServer) NodePublishVolume(ctx context.Context, req *csi.NodePublishVolumeRequest) (*csi.NodePublishVolumeResponse, error) {
	klog.V(4).InfoS("Begin to publish volume")

	// Step 0: check the necessary parameters
	volumeID := req.GetVolumeId()
	if volumeID == "" {
		return nil, status.Error(codes.InvalidArgument, "Volume ID is required")
	}

	targetPath := req.GetTargetPath()
	if targetPath == "" {
		return nil, status.Error(codes.InvalidArgument, "Target path is required")
	}

	volumeCapability := req.GetVolumeCapability()
	if volumeCapability == nil {
		return nil, status.Error(codes.InvalidArgument, "Volume capability is required")
	}

	mountOpts := volumeCapability.GetMount().GetMountFlags()
	if req.GetReadonly() {
		mountOpts = append(mountOpts, "ro")
	}

	var server, basedir string
	volumeContext := req.GetVolumeContext()

	mountPermissionsValue := volumeContext[mountPermissionKey]
	var (
		mountPermissions uint64
		err              error
	)
	if mountPermissions, err = strconv.ParseUint(mountPermissionsValue, 8, 32); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, err.Error())
	}

	server = volumeContext[serverKey]
	if server == "" {
		return nil, status.Error(codes.InvalidArgument, "Server is required")
	}

	basedir = volumeContext[basedirKey]
	if basedir == "" {
		return nil, status.Error(codes.InvalidArgument, "Base directory is required")
	}
	basedir = strings.Trim(basedir, string(filepath.Separator))
	basedir = filepath.Join(string(filepath.Separator), basedir)
	source := fmt.Sprintf("%s:%s", server, basedir)

	/* subdir = volumeContext[subdirKey]
	if subdir == "" {
		return nil, status.Error(codes.InvalidArgument, "Sub directory is required")
	} */

	notMnt, err := ns.mounter.IsLikelyNotMountPoint(targetPath)
	if err != nil {
		if os.IsNotExist(err) {
			if err := os.MkdirAll(targetPath, os.FileMode(mountPermissions)); err != nil {
				return nil, status.Error(codes.Internal, err.Error())
			}
			notMnt = true
		} else {
			return nil, status.Error(codes.Internal, err.Error())
		}
	}
	if !notMnt {
		return &csi.NodePublishVolumeResponse{}, nil
	}

	// Step 1: do mount
	klog.V(4).Infof("NodePublishVolume: volumeID(%v) source(%s) targetPath(%s) mountflags(%v)", volumeID, source, targetPath, mountOpts)
	err = ns.mounter.Mount(source, targetPath, "nfs", mountOpts)
	if err != nil {
		if os.IsPermission(err) {
			return nil, status.Errorf(codes.PermissionDenied, err.Error())
		}
		if strings.Contains(err.Error(), "invalid argument") {
			return nil, status.Errorf(codes.InvalidArgument, "invalid argument: %v", err)
		}
		return nil, status.Errorf(codes.Internal, "mount failed: %v", err)
	}

	// Step 2: check the rightness of the mount result
	if mountPermissions > 0 {
		if err := checkMountPermissions(targetPath, os.FileMode(mountPermissions)); err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}
	} else {
		klog.V(4).InfoS("No validate since it is 0")
	}

	klog.V(4).InfoS("Mount succeeded")
	return &csi.NodePublishVolumeResponse{}, nil
}

func (ns *nodeServer) NodeUnpublishVolume(ctx context.Context, req *csi.NodeUnpublishVolumeRequest) (*csi.NodeUnpublishVolumeResponse, error) {
	volId := req.GetVolumeId()
	if volId == "" {
		return nil, status.Error(codes.InvalidArgument, "Volume ID is required")
	}

	targetPath := req.GetTargetPath()
	if targetPath == "" {
		return nil, status.Error(codes.InvalidArgument, "Target path is required")
	}

	if err := mount.CleanupMountPoint(targetPath, ns.mounter, false); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to unmount %s: %v", targetPath, err.Error())
	}

	return &csi.NodeUnpublishVolumeResponse{}, nil
}

// NodeGetCapabilities implements csi.NodeServer.
func (ns *nodeServer) NodeGetCapabilities(context.Context, *csi.NodeGetCapabilitiesRequest) (*csi.NodeGetCapabilitiesResponse, error) {
	return &csi.NodeGetCapabilitiesResponse{
		Capabilities: ns.driver.nodeCaps,
	}, nil
}

// NodeGetInfo implements csi.NodeServer.
func (ns *nodeServer) NodeGetInfo(context.Context, *csi.NodeGetInfoRequest) (*csi.NodeGetInfoResponse, error) {
	return &csi.NodeGetInfoResponse{
		NodeId: ns.driver.node,
	}, nil
}

// NodeGetVolumeStats implements csi.NodeServer.
func (ns *nodeServer) NodeGetVolumeStats(ctx context.Context, req *csi.NodeGetVolumeStatsRequest) (*csi.NodeGetVolumeStatsResponse, error) {
	if req.VolumeId == "" {
		return nil, status.Error(codes.InvalidArgument, "Volume ID is required")
	}

	targetPath := req.GetVolumePath()
	if targetPath == "" {
		return nil, status.Error(codes.InvalidArgument, "Volume path is required")
	}

	if _, err := os.Lstat(targetPath); err != nil {
		if os.IsNotExist(err) {
			return nil, status.Errorf(codes.NotFound, "Volume path not found: %s", targetPath)
		}
		return nil, status.Errorf(codes.Internal, "failed to stat volume path %s: %v", targetPath, err)
	}

	// Generate stats
	volumeMetrics, err := volume.NewMetricsStatFS(req.VolumePath).GetMetrics()
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get metrics: %v", err)
	}

	available, ok := volumeMetrics.Available.AsInt64()
	if !ok {
		return nil, status.Errorf(codes.Internal, "failed to transform volume available size(%v)", volumeMetrics.Available)
	}
	capacity, ok := volumeMetrics.Capacity.AsInt64()
	if !ok {
		return nil, status.Errorf(codes.Internal, "failed to transform volume capacity size(%v)", volumeMetrics.Capacity)
	}
	used, ok := volumeMetrics.Used.AsInt64()
	if !ok {
		return nil, status.Errorf(codes.Internal, "failed to transform volume used size(%v)", volumeMetrics.Used)
	}

	inodesFree, ok := volumeMetrics.InodesFree.AsInt64()
	if !ok {
		return nil, status.Errorf(codes.Internal, "failed to transform disk inodes free(%v)", volumeMetrics.InodesFree)
	}
	inodes, ok := volumeMetrics.Inodes.AsInt64()
	if !ok {
		return nil, status.Errorf(codes.Internal, "failed to transform disk inodes(%v)", volumeMetrics.Inodes)
	}
	inodesUsed, ok := volumeMetrics.InodesUsed.AsInt64()
	if !ok {
		return nil, status.Errorf(codes.Internal, "failed to transform disk inodes used(%v)", volumeMetrics.InodesUsed)
	}

	return &csi.NodeGetVolumeStatsResponse{
		Usage: []*csi.VolumeUsage{
			{
				Unit:      csi.VolumeUsage_BYTES,
				Available: available,
				Total:     capacity,
				Used:      used,
			},
			{
				Unit:      csi.VolumeUsage_INODES,
				Available: inodesFree,
				Total:     inodes,
				Used:      inodesUsed,
			},
		},
	}, nil
}

// NodeStageVolume implements csi.NodeServer.
func (ns *nodeServer) NodeStageVolume(context.Context, *csi.NodeStageVolumeRequest) (*csi.NodeStageVolumeResponse, error) {
	return nil, status.Error(codes.Unimplemented, "NodeStageVolume is not implemented")
}

// NodeUnstageVolume implements csi.NodeServer.
func (ns *nodeServer) NodeUnstageVolume(context.Context, *csi.NodeUnstageVolumeRequest) (*csi.NodeUnstageVolumeResponse, error) {
	return nil, status.Error(codes.Unimplemented, "NodeUnstageVolume is not implemented")
}

func (ns *nodeServer) NodeExpandVolume(ctx context.Context, req *csi.NodeExpandVolumeRequest) (*csi.NodeExpandVolumeResponse, error) {
	return nil, status.Error(codes.Unimplemented, "NodeExpandVolume is not implemented")
}

func checkMountPermissions(targetPath string, mode os.FileMode) error {
	info, err := os.Lstat(targetPath)
	if err != nil {
		return err
	}

	perm := info.Mode().Perm()
	if perm != mode {
		klog.V(4).InfoS("mode mismatch, changing...")
		if err := os.Chmod(targetPath, mode); err != nil {
			return err
		}
	} else {
		klog.V(4).InfoS("mode match")
	}
	return nil
}
