package nfs

import (
	"context"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	csi "github.com/container-storage-interface/spec/lib/go/csi"
	mount "k8s.io/mount-utils"
)

var (
	testTargetPath = filepath.Join(os.TempDir(), "testTargetPath")
)

const (
	testVolId = "testVolId"

	testValidateMountPermissions = "0777"
	testServer                   = "testServer"
	testBasePath                 = "testBasePath"
	testSubPath                  = "testSubPath"
)

func TestValidateMountPermissions(t *testing.T) {
	type args struct {
		targetPath string
		mode       os.FileMode
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Invalid permissions",
			args: args{
				targetPath: "/tmp",
				mode:       0777,
			},
			wantErr: true,
		},
		{
			name: "Valid permissions",
			args: args{
				targetPath: "/tmp",
				mode:       0755,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := checkMountPermissions(tt.args.targetPath, tt.args.mode); (err != nil) != tt.wantErr {
				t.Errorf("validateMountPermissions() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestNodeGetInfo(t *testing.T) {
	type fields struct {
		driver *nfsDriver
	}
	type args struct {
		in0 context.Context
		in1 *csi.NodeGetInfoRequest
	}
	tests := []struct {
		name     string
		fields   fields
		args     args
		want     *csi.NodeGetInfoResponse
		expected bool
	}{
		{
			name: "Right NodeGetInfo",
			fields: fields{
				driver: NewFakeNfsDriver(fakeNode),
			},
			args: args{
				in0: context.Background(),
				in1: &csi.NodeGetInfoRequest{},
			},
			want: &csi.NodeGetInfoResponse{
				NodeId: fakeNode,
			},
			expected: true,
		},
		{
			name: "Wrong NodeGetInfo",
			fields: fields{
				driver: NewFakeNfsDriver("wrongNode"),
			},
			args: args{
				in0: context.Background(),
				in1: &csi.NodeGetInfoRequest{},
			},
			want: &csi.NodeGetInfoResponse{
				NodeId: fakeNode,
			},
			expected: false,
		},
	}
	for _, tt := range tests {
		ns := &nodeServer{
			driver: tt.fields.driver,
		}
		t.Run(tt.name, func(t *testing.T) {
			got, _ := ns.NodeGetInfo(tt.args.in0, tt.args.in1)
			if !reflect.DeepEqual(tt.expected, (got.NodeId == tt.want.NodeId)) {
				t.Errorf("nodeServer.NodeGetInfo() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNodePublishVolume(t *testing.T) {
	type fields struct {
		driver  *nfsDriver
		mounter mount.Interface
	}
	type args struct {
		ctx context.Context
		req *csi.NodePublishVolumeRequest
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *csi.NodePublishVolumeResponse
		wantErr bool
	}{
		{
			name: "Empty Volume ID",
			fields: fields{
				driver:  NewFakeNfsDriver(fakeNode),
				mounter: mount.NewFakeMounter([]mount.MountPoint{}),
			},
			args: args{
				ctx: context.Background(),
				req: &csi.NodePublishVolumeRequest{
					VolumeId: "",
				},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "Empty Target Path",
			fields: fields{
				driver:  NewFakeNfsDriver(fakeNode),
				mounter: mount.NewFakeMounter([]mount.MountPoint{}),
			},
			args: args{
				ctx: context.Background(),
				req: &csi.NodePublishVolumeRequest{
					VolumeId:   testVolId,
					TargetPath: "",
				},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "Empty Volume Capabilities",
			fields: fields{
				driver:  NewFakeNfsDriver(fakeNode),
				mounter: mount.NewFakeMounter([]mount.MountPoint{}),
			},
			args: args{
				ctx: context.Background(),
				req: &csi.NodePublishVolumeRequest{
					VolumeId:   testVolId,
					TargetPath: testTargetPath,
				},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "Wrong Mount Permissions",
			fields: fields{
				driver:  NewFakeNfsDriver(fakeNode),
				mounter: mount.NewFakeMounter([]mount.MountPoint{}),
			},
			args: args{
				ctx: context.Background(),
				req: &csi.NodePublishVolumeRequest{
					VolumeId:   testVolId,
					TargetPath: testTargetPath,
					VolumeCapability: &csi.VolumeCapability{
						AccessType: &csi.VolumeCapability_Mount{},
					},
					VolumeContext: map[string]string{
						"mountPermissions": "07cc",
					},
				},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "Empty server",
			fields: fields{
				driver:  NewFakeNfsDriver(fakeNode),
				mounter: mount.NewFakeMounter([]mount.MountPoint{}),
			},
			args: args{
				ctx: context.Background(),
				req: &csi.NodePublishVolumeRequest{
					VolumeId:   testVolId,
					TargetPath: testTargetPath,
					VolumeCapability: &csi.VolumeCapability{
						AccessType: &csi.VolumeCapability_Mount{},
					},
					VolumeContext: map[string]string{
						"mountPermissions": testValidateMountPermissions,
					},
				},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "Empty basedir",
			fields: fields{
				driver:  NewFakeNfsDriver(fakeNode),
				mounter: mount.NewFakeMounter([]mount.MountPoint{}),
			},
			args: args{
				ctx: context.Background(),
				req: &csi.NodePublishVolumeRequest{
					VolumeId:   testVolId,
					TargetPath: testTargetPath,
					VolumeCapability: &csi.VolumeCapability{
						AccessType: &csi.VolumeCapability_Mount{},
					},
					VolumeContext: map[string]string{
						"mountPermissions": testValidateMountPermissions,
						"server":           testServer,
					},
				},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "Empty basedir",
			fields: fields{
				driver:  NewFakeNfsDriver(fakeNode),
				mounter: mount.NewFakeMounter([]mount.MountPoint{}),
			},
			args: args{
				ctx: context.Background(),
				req: &csi.NodePublishVolumeRequest{
					VolumeId:   testVolId,
					TargetPath: testTargetPath,
					VolumeCapability: &csi.VolumeCapability{
						AccessType: &csi.VolumeCapability_Mount{},
					},
					VolumeContext: map[string]string{
						"mountPermissions": testValidateMountPermissions,
						"server":           testServer,
						"basedir":          testBasePath,
					},
				},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "Wrong subpath",
			fields: fields{
				driver:  NewFakeNfsDriver(fakeNode),
				mounter: mount.NewFakeMounter([]mount.MountPoint{}),
			},
			args: args{
				ctx: context.Background(),
				req: &csi.NodePublishVolumeRequest{
					VolumeId:   testVolId,
					TargetPath: testTargetPath,
					VolumeCapability: &csi.VolumeCapability{
						AccessType: &csi.VolumeCapability_Mount{},
					},
					VolumeContext: map[string]string{
						"mountPermissions": testValidateMountPermissions,
						"server":           testServer,
						"basedir":          testBasePath,
						"subdir":           testSubPath,
					},
				},
			},
			want:    &csi.NodePublishVolumeResponse{},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ns := &nodeServer{
				driver:  tt.fields.driver,
				mounter: tt.fields.mounter,
			}
			got, err := ns.NodePublishVolume(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("nodeServer.NodePublishVolume() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("nodeServer.NodePublishVolume() = %v, want %v", got, tt.want)
			}
		})
	}
	if err := os.RemoveAll(testTargetPath); err != nil {
		t.Errorf("Failed to remove test target path %s", testTargetPath)
	}
}

func TestNodeUnpublishVolume(t *testing.T) {
	if err := os.MkdirAll(testTargetPath, 0750); err != nil {
		t.Errorf("Failed to create test target path %s", testTargetPath)
	}

	type fields struct {
		driver  *nfsDriver
		mounter mount.Interface
	}
	type args struct {
		ctx context.Context
		req *csi.NodeUnpublishVolumeRequest
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *csi.NodeUnpublishVolumeResponse
		wantErr bool
	}{
		{
			name: "Empty volume id",
			fields: fields{
				driver:  NewFakeNfsDriver(fakeNode),
				mounter: mount.NewFakeMounter([]mount.MountPoint{}),
			},
			args: args{
				ctx: context.Background(),
				req: &csi.NodeUnpublishVolumeRequest{},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "Empty target path",
			fields: fields{
				driver:  NewFakeNfsDriver(fakeNode),
				mounter: mount.NewFakeMounter([]mount.MountPoint{}),
			},
			args: args{
				ctx: context.Background(),
				req: &csi.NodeUnpublishVolumeRequest{
					VolumeId: testVolId,
				},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "Successful Unpublish",
			fields: fields{
				driver:  NewFakeNfsDriver(fakeNode),
				mounter: mount.NewFakeMounter([]mount.MountPoint{}),
			},
			args: args{
				ctx: context.Background(),
				req: &csi.NodeUnpublishVolumeRequest{
					VolumeId:   testVolId,
					TargetPath: testTargetPath,
				},
			},
			want:    &csi.NodeUnpublishVolumeResponse{},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ns := &nodeServer{
				driver:  tt.fields.driver,
				mounter: tt.fields.mounter,
			}
			got, err := ns.NodeUnpublishVolume(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("nodeServer.NodeUnpublishVolume() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("nodeServer.NodeUnpublishVolume() = %v, want %v", got, tt.want)
			}
		})
	}
}
