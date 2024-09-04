package types

import (
	"reflect"
	"testing"
)

func TestGetTargetManifest(t *testing.T) {
	targetManifest := GetTargetManifest()
	if targetManifest == nil {
		t.Fatalf("Expected target manifest but got nil")
	}

	fields := [8]string{"Region", "Image Id", "Instance Type", "Device Name",
		"Volume Size", "Volume Type", "Access Key Id", "Secret Access Key",
	}
	for _, field := range fields {
		if _, ok := (*targetManifest)[field]; !ok {
			t.Errorf("Expected field %s in target manifest but it was not found", field)
		}
	}
}

func TestParseTargetOptions(t *testing.T) {
	tests := []struct {
		name        string
		optionsJson string
		envVars     map[string]string
		want        *TargetOptions
		wantErr     bool
	}{
		{
			name: "Valid JSON with all fields",
			optionsJson: `{
				"Region": "us-west-2",
				"Image Id": "ami-12345678",
				"Instance Type": "t3.micro",
				"Device Name": "/dev/sda1",
				"Volume Size": 20,
				"Volume Type": "gp2",
				"Access Key Id": "accessKeyID",
				"Secret Access Key": "secretAccessKey"
			}`,
			want: &TargetOptions{
				Region:          "us-west-2",
				ImageId:         "ami-12345678",
				InstanceType:    "t3.micro",
				DeviceName:      "/dev/sda1",
				VolumeSize:      20,
				VolumeType:      "gp2",
				AccessKeyId:     "accessKeyID",
				SecretAccessKey: "secretAccessKey",
			},
			wantErr: false,
		},
		{
			name: "Valid JSON with missing credentials, using env vars",
			optionsJson: `{
				"Region": "us-east-1",
				"Image Id": "ami-87654321",
				"Instance Type": "t2.micro",
				"Device Name": "/dev/xvda",
				"Volume Size": 10,
				"Volume Type": "gp3"
			}`,
			envVars: map[string]string{
				"AWS_ACCESS_KEY_ID":     "accessKeyID",
				"AWS_SECRET_ACCESS_KEY": "secretAccessKey",
				"AWS_DEFAULT_REGION":    "us-east-1",
			},
			want: &TargetOptions{
				Region:          "us-east-1",
				ImageId:         "ami-87654321",
				InstanceType:    "t2.micro",
				DeviceName:      "/dev/xvda",
				VolumeSize:      10,
				VolumeType:      "gp3",
				AccessKeyId:     "accessKeyID",
				SecretAccessKey: "secretAccessKey",
			},
			wantErr: false,
		},
		{
			name:        "Invalid JSON",
			optionsJson: `{"Region": "us-east-1", "Image ID": "ami-12345678"`,
			wantErr:     true,
		},
		{
			name: "Missing credentials in both JSON and env vars",
			optionsJson: `{
				"Region": "us-east-1",
				"Image ID": "ami-12345678"
			}`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for k, v := range tt.envVars {
				t.Setenv(k, v)
			}

			got, err := ParseTargetOptions(tt.optionsJson)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseTargetOptions() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParseTargetOptions() = %v, want %v", got, tt.want)
			}
		})
	}
}
