package nfs

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/chenliu1993/simple-csi-driver/internal/idempotency"
	csi "github.com/container-storage-interface/spec/lib/go/csi"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"k8s.io/klog/v2"
)

// TODO: RetainPolicy

// Check if implements csi.ControllerServer
var _ csi.ControllerServer = &controllerServer{}

type controllerServer struct {
	driver *nfsDriver

	idempotency *idempotency.Idempotency
}

func NewControllerServer(driver *nfsDriver) *controllerServer {
	return &controllerServer{
		driver: driver,

		idempotency: idempotency.NewIdempotency(),
	}
}

// CreateVolume creates a nfs-type volume
func (cs *controllerServer) CreateVolume(ctx context.Context, req *csi.CreateVolumeRequest) (*csi.CreateVolumeResponse, error) {
	klog.V(4).InfoS("Creating volume......")

	// Step 10: validate thr request parameters
	if err := validateVolumeRequest(req); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	// Step 1: check if the volume is being handled
	if cs.idempotency.IsProcessing(req.Name) {
		return nil, status.Error(codes.Aborted, "Volume is being handled")
	}
	cs.idempotency.AddProcessing(req.Name)

	parameters := req.GetParameters()

	// Step 2: create the volume
	if _, ok := parameters[subdirKey]; !ok || parameters[subdirKey] == "" {
		parameters[subdirKey] = req.GetName()
	}
	volId := getVolIdFromParams(parameters)

	targetParentPath := getTargetParentPath(parameters[subdirKey])
	if err := cs.preMount(ctx, parameters, volId, targetParentPath); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	// Needs to unmoiunt since we are just creating the volume, not to publish it them
	defer func() {
		klog.V(4).InfoS("Unmounting at target path: ", targetParentPath)
		if err := cs.preUnmount(ctx, volId, targetParentPath); err != nil {
			klog.Warningf("failed to unmount nfs server: %v", err)
		}
		cs.idempotency.RemoveProcessing(req.Name)
	}()

	// Step 3: Create the actual target path
	mountPermission, err := strconv.ParseUint(parameters[mountPermissionKey], 8, 32)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	volumeMountPath := getVolumtMountPath(targetParentPath, parameters[subdirKey])
	if err := os.MkdirAll(volumeMountPath, os.FileMode(mountPermission)); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &csi.CreateVolumeResponse{
		Volume: &csi.Volume{
			VolumeId:      volId,
			CapacityBytes: 0,
			VolumeContext: parameters,
		},
	}, nil
}

// DeleteVolume deletes a nfs-type volume
func (cs *controllerServer) DeleteVolume(ctx context.Context, req *csi.DeleteVolumeRequest) (*csi.DeleteVolumeResponse, error) {
	klog.V(4).InfoS("Deleting volume......")

	// Step 0: check if the volume is being handled
	if cs.idempotency.IsProcessing(req.VolumeId) {
		return nil, status.Error(codes.FailedPrecondition, "Volume is gone already")
	}
	cs.idempotency.AddProcessing(req.VolumeId)

	// step 1: simple check
	volId := req.VolumeId
	if volId == "" {
		return nil, status.Error(codes.InvalidArgument, "Volume ID is required")
	}
	server, basedir, subdir, err := getParamsFromVolId(volId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	parameters := map[string]string{}
	parameters[serverKey] = server
	parameters[basedirKey] = basedir
	parameters[subdirKey] = subdir

	// step 2: delete the volume target path,
	// Since when DeleteVolume is called, the volume is no longer attached,
	// Thus remount again
	targetParentPath := getTargetParentPath(parameters[subdirKey])
	if err := cs.preMount(ctx, parameters, volId, targetParentPath); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	// Needs to unmoiunt since we are just creating the volume, not to publish it them
	defer func() {
		klog.V(4).InfoS("Unmounting at target path: ", targetParentPath)
		if err := cs.preUnmount(ctx, volId, targetParentPath); err != nil {
			klog.Warningf("failed to unmount nfs server: %v", err)
		}
		cs.idempotency.RemoveProcessing(req.VolumeId)
	}()

	volumeMountPath := getVolumtMountPath(targetParentPath, parameters[subdirKey])
	klog.V(4).InfoS("Removing the actual volume path: ", volumeMountPath)
	if err := os.RemoveAll(volumeMountPath); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	cs.idempotency.RemoveProcessing(req.VolumeId)
	return &csi.DeleteVolumeResponse{}, nil
}

func (cs *controllerServer) ValidateVolumeCapabilities(ctx context.Context, req *csi.ValidateVolumeCapabilitiesRequest) (*csi.ValidateVolumeCapabilitiesResponse, error) {
	volId := req.GetVolumeId()
	if volId == "" {
		return nil, status.Error(codes.InvalidArgument, "Volume ID is required")
	}

	if err := tryValidateVolumeCapabilities(req.GetVolumeCapabilities()); err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid volume capabilities")
	}

	return &csi.ValidateVolumeCapabilitiesResponse{
		Confirmed: &csi.ValidateVolumeCapabilitiesResponse_Confirmed{
			VolumeCapabilities: req.GetVolumeCapabilities(),
		},
		Message: "",
	}, nil
}

func (cs *controllerServer) ControllerGetCapabilities(ctx context.Context, req *csi.ControllerGetCapabilitiesRequest) (*csi.ControllerGetCapabilitiesResponse, error) {
	return &csi.ControllerGetCapabilitiesResponse{
		Capabilities: cs.driver.controllerCaps,
	}, nil
}

// Below are unimplemented functions

func (cs *controllerServer) CreateSnapshot(ctx context.Context, req *csi.CreateSnapshotRequest) (*csi.CreateSnapshotResponse, error) {
	return nil, status.Error(codes.Unimplemented, "Unimplemented")
}

func (cs *controllerServer) DeleteSnapshot(ctx context.Context, req *csi.DeleteSnapshotRequest) (*csi.DeleteSnapshotResponse, error) {
	return nil, status.Error(codes.Unimplemented, "Unimplemented")
}

func (cs *controllerServer) ListSnapshots(ctx context.Context, req *csi.ListSnapshotsRequest) (*csi.ListSnapshotsResponse, error) {
	return nil, status.Error(codes.Unimplemented, "")
}

// ControllerPublishVolume attaches a volume to a node VM
func (cs *controllerServer) ControllerPublishVolume(ctx context.Context, req *csi.ControllerPublishVolumeRequest) (*csi.ControllerPublishVolumeResponse, error) {
	klog.V(4).InfoS("NFS doesn't need publishing volume to node, skipping......")
	return nil, status.Error(codes.Unimplemented, "Unimplemented")
}

// ControllerPublishVolume detaches a volume from a node VM
func (cs *controllerServer) ControllerUnpublishVolume(ctx context.Context, req *csi.ControllerUnpublishVolumeRequest) (*csi.ControllerUnpublishVolumeResponse, error) {
	klog.V(4).InfoS("NFS doesn't need unpublishing volume to node, skipping......")

	return nil, status.Error(codes.Unimplemented, "Unimplemented")
}

func (cs *controllerServer) ControllerExpandVolume(ctx context.Context, req *csi.ControllerExpandVolumeRequest) (*csi.ControllerExpandVolumeResponse, error) {
	klog.V(4).InfoS("NFS doesn't need expanding volume, skipping......")

	return nil, status.Error(codes.Unimplemented, "Unimplemented")
}

func (cs *controllerServer) GetCapacity(ctx context.Context, req *csi.GetCapacityRequest) (*csi.GetCapacityResponse, error) {
	return nil, status.Error(codes.Unimplemented, "")
}

func (cs *controllerServer) ControllerGetVolume(ctx context.Context, req *csi.ControllerGetVolumeRequest) (*csi.ControllerGetVolumeResponse, error) {
	return nil, status.Error(codes.Unimplemented, "Unimplemented")
}

func (cs *controllerServer) ListVolumes(ctx context.Context, req *csi.ListVolumesRequest) (*csi.ListVolumesResponse, error) {
	return nil, status.Error(codes.Unimplemented, "Unimplemented")
}

func validateVolumeRequest(req *csi.CreateVolumeRequest) error {
	klog.V(4).InfoS("Validating volume request parameters......")

	if len(req.GetName()) == 0 {
		return errors.New("volume name cannot be empty")
	}

	parameters := req.GetParameters()
	if parameters == nil {
		req.Parameters = make(map[string]string)
	}

	return validateNfsParameters(req.Parameters)
}

func validateNfsParameters(params map[string]string) error {
	var msg string
	for _, value := range params {
		if strings.Contains(value, "&") || strings.Contains(value, "|") || strings.Contains(value, ";") ||
			strings.Contains(value, "$") || strings.Contains(value, "'") || strings.Contains(value, "`") ||
			strings.Contains(value, "(") || strings.Contains(value, ")") {
			msg = msg + fmt.Sprintf("Args %s has illegal access.", value)
		}
	}
	if msg != "" {
		return errors.New(msg)
	}

	if err := validateMountPermissions(params[mountPermissionKey]); err != nil {
		return err
	}

	if _, ok := params[serverKey]; !ok {
		return errors.New("nfs server is required")
	}
	if _, ok := params[basedirKey]; !ok {
		return errors.New("nfs basedir is required")
	}
	/* if _, ok := params[subdirKey]; !ok {
		return errors.New("Nfs subdir is required")
	} */

	return nil
}

// getVolumtMountPath returns the path where the volume will be mounted
func getVolumtMountPath(targetParentPath string, subdir string) string {
	return filepath.Join(targetParentPath, subdir)
}

// getVolI generates a unique volume ID based on the parameters
func getVolIdFromParams(parameters map[string]string) string {
	server := strings.Trim(parameters[serverKey], "/")
	basedir := strings.Trim(parameters[basedirKey], "/")
	subDir := strings.Trim(parameters[subdirKey], "/")
	volIdElements := []string{
		server,
		basedir,
		subDir,
	}
	return strings.Join(volIdElements, seperator)
}

func getParamsFromVolId(volId string) (string, string, string, error) {
	volIdElements := strings.Split(volId, seperator)
	if len(volIdElements) != 3 {
		return "", "", "", errors.New("invalid volume ID which cannot be parsed")
	}
	return volIdElements[0], volIdElements[1], volIdElements[2], nil
}

// getTargetPath returns the shared path of the nfs server, nfs source folder will be created under this path
func getTargetParentPath(mountPath string) string {
	return filepath.Join(nfsWorkingDir, mountPath)
}

func (cs *controllerServer) preMount(ctx context.Context, parameters map[string]string, volId, targetParentPath string) error {
	klog.V(4).InfoS("do a premounting, create the subdir in advance")

	volCap := &csi.VolumeCapability{
		AccessType: &csi.VolumeCapability_Mount{
			Mount: &csi.VolumeCapability_MountVolume{},
		},
	}

	server := strings.Trim(parameters[serverKey], "/")
	basedir := strings.Trim(parameters[basedirKey], "/")
	volumeContext := map[string]string{
		serverKey:          server,
		basedirKey:         filepath.Join(string(filepath.Separator), basedir),
		mountPermissionKey: parameters[mountPermissionKey],
	}
	for k, v := range parameters {
		if strings.ToLower(k) != subdirKey {
			volumeContext[k] = v
		}
	}

	// In this step, we only mount server:/basedir
	if _, err := cs.driver.ns.NodePublishVolume(ctx, &csi.NodePublishVolumeRequest{
		TargetPath:       targetParentPath,
		VolumeContext:    volumeContext,
		VolumeCapability: volCap,
		VolumeId:         volId,
	}); err != nil {
		return err
	}
	return nil
}

func (cs *controllerServer) preUnmount(ctx context.Context, volId, targetParentPath string) error {
	klog.V(4).InfoS("do a unmount since this is just pre")
	if _, err := cs.driver.ns.NodeUnpublishVolume(ctx, &csi.NodeUnpublishVolumeRequest{
		TargetPath: targetParentPath,
		VolumeId:   volId,
	}); err != nil {
		return err
	}
	return nil
}

func tryValidateVolumeCapabilities(volCaps []*csi.VolumeCapability) error {
	if len(volCaps) == 0 {
		return errors.New("volume capabilities cannot be empty")
	}

	// just check if the unsupported type is inclueded
	for _, cap := range volCaps {
		if cap.GetBlock() != nil {
			return errors.New("block volume is not supported")
		}
	}
	return nil
}

func validateMountPermissions(mountPermissions string) error {
	if _, err := strconv.ParseUint(mountPermissions, 8, 32); err != nil {
		return fmt.Errorf("invalid mount permissions: %s", mountPermissions)
	}
	return nil
}
