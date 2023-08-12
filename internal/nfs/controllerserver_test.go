package nfs

import (
	"context"
	"reflect"
	"testing"

	"github.com/chenliu1993/simple-csi-driver/internal/idempotency"
	csi "github.com/container-storage-interface/spec/lib/go/csi"
)

func TestCreateVolume(t *testing.T) {
	type fields struct {
		driver      *nfsDriver
		idempotency *idempotency.Idempotency
	}
	type args struct {
		ctx context.Context
		req *csi.CreateVolumeRequest
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "create volume with problematic parameters",
			fields: fields{
				driver:      NewFakeNfsDriver(fakeNode),
				idempotency: idempotency.NewIdempotency(),
			},
			args: args{
				ctx: context.Background(),
				req: &csi.CreateVolumeRequest{
					Name: "testCreateVolumeReq1",
					Parameters: map[string]string{
						"unknown": "1&1(1)1",
					},
				},
			},
			wantErr: true,
		},
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cs := &controllerServer{
				driver:      tt.fields.driver,
				idempotency: tt.fields.idempotency,
			}
			_, err := cs.CreateVolume(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("controllerServer.CreateVolume() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

func TestDeleteVolume(t *testing.T) {
	type fields struct {
		driver      *nfsDriver
		idempotency *idempotency.Idempotency
	}
	type args struct {
		ctx context.Context
		req *csi.DeleteVolumeRequest
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *csi.DeleteVolumeResponse
		wantErr bool
	}{
		{
			name: "delete volume with problematic volId",
			fields: fields{
				driver:      NewFakeNfsDriver(fakeNode),
				idempotency: idempotency.NewIdempotency(),
			},
			args: args{
				ctx: context.Background(),
				req: &csi.DeleteVolumeRequest{
					VolumeId: "testDeleteVolumeReq1",
				},
			},
			wantErr: true,
		},
		// TODO Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cs := &controllerServer{
				driver:      tt.fields.driver,
				idempotency: tt.fields.idempotency,
			}
			got, err := cs.DeleteVolume(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("controllerServer.DeleteVolume() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("controllerServer.DeleteVolume() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestControllerGetCapabilities(t *testing.T) {
	type fields struct {
		driver      *nfsDriver
		idempotency *idempotency.Idempotency
	}
	type args struct {
		ctx context.Context
		req *csi.ControllerGetCapabilitiesRequest
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *csi.ControllerGetCapabilitiesResponse
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cs := &controllerServer{
				driver:      tt.fields.driver,
				idempotency: tt.fields.idempotency,
			}
			got, err := cs.ControllerGetCapabilities(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("controllerServer.ControllerGetCapabilities() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("controllerServer.ControllerGetCapabilities() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetVolumtMountPath(t *testing.T) {
	type args struct {
		targetParentPath string
		subdir           string
	}
	tests := []struct {
		name         string
		args         args
		expectedPath string
	}{
		{
			name: "get volume mount path",
			args: args{
				targetParentPath: "/tmp",
				subdir:           "fakeVol1",
			},
			expectedPath: "/tmp/fakeVol1",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if actualPath := getVolumtMountPath(tt.args.targetParentPath, tt.args.subdir); actualPath != tt.expectedPath {
				t.Errorf("getVolumtMountPath() = %v, want %v", actualPath, tt.expectedPath)
			}
		})
	}
}

func TestGetVolIdFromParams(t *testing.T) {
	type args struct {
		parameters map[string]string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "get volume id from parameters",
			args: args{
				parameters: map[string]string{
					"server":  "faleServer",
					"basedir": "fakeBaseDir",
					"subdir":  "fakeSubDir",
				},
			},
			want: "faleServer#fakeBaseDir#fakeSubDir",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getVolIdFromParams(tt.args.parameters); got != tt.want {
				t.Errorf("getVolIdFromParams() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetTargetParentPath(t *testing.T) {
	type args struct {
		volId string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{

			name: "get target parent path",
			args: args{
				volId: "faleServer#fakeBaseDir#fakeSubDir",
			},
			want: "/tmp/faleServer#fakeBaseDir#fakeSubDir",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getTargetParentPath(tt.args.volId); got != tt.want {
				t.Errorf("getTargetParentPath() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetParamsFromVolId(t *testing.T) {
	type args struct {
		volId string
	}
	tests := []struct {
		name  string
		args  args
		want  string
		want1 string
		want2 string
	}{
		{
			name: "get params from volume id",
			args: args{
				volId: "faleServer#fakeBaseDir#fakeSubDir",
			},
			want:  "faleServer",
			want1: "fakeBaseDir",
			want2: "fakeSubDir",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1, got2, err := getParamsFromVolId(tt.args.volId)
			if got != tt.want {
				t.Errorf("getParamsFromVolId() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("getParamsFromVolId() got1 = %v, want %v", got1, tt.want1)
			}
			if got2 != tt.want2 {
				t.Errorf("getParamsFromVolId() got2 = %v, want %v", got2, tt.want2)
			}
			if err != nil {
				t.Errorf("getParamsFromVolId() error = %v", err)
			}
		})
	}
}

func TestValidateNfsParameters(t *testing.T) {
	type args struct {
		params map[string]string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "validate illegal parameters",
			args: args{
				params: map[string]string{
					"server": "&",
				},
			},
			wantErr: true,
		},
		{
			name: "validate empty server",
			args: args{
				params: map[string]string{},
			},
			wantErr: true,
		},
		{
			name: "validate empty basedir",
			args: args{
				params: map[string]string{
					"server": "fakeServer",
				},
			},
			wantErr: true,
		},
		{
			name: "validate empty subdir",
			args: args{
				params: map[string]string{
					"server":  "fakeServer",
					"basedir": "fakeBaseDir",
				},
			},
			wantErr: true,
		},
		{
			name: "validate rightParameters",
			args: args{
				params: map[string]string{
					"server":  "fakeServer",
					"basedir": "fakeBaseDir",
					"subdir":  "fakeSubDir",
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := validateNfsParameters(tt.args.params); (err != nil) != tt.wantErr {
				t.Errorf("validateNfsParameters() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestTryValidateVolumeCapabilities(t *testing.T) {
	type args struct {
		volCaps []*csi.VolumeCapability
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{

		{
			name:    "empty volume capabilities",
			args:    args{},
			wantErr: true,
		},
		{
			name: "block volume capabilities",
			args: args{
				volCaps: []*csi.VolumeCapability{
					{
						AccessType: &csi.VolumeCapability_Block{
							Block: &csi.VolumeCapability_BlockVolume{},
						},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "volume capabilities without problem",
			args: args{
				volCaps: []*csi.VolumeCapability{
					{
						AccessType: &csi.VolumeCapability_Mount{
							Mount: &csi.VolumeCapability_MountVolume{},
						},
					},
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tryValidateVolumeCapabilities(tt.args.volCaps); (err != nil) != tt.wantErr {
				t.Errorf("tryValidateVolumeCapabilities() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
