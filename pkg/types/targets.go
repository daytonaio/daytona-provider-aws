package types

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/daytonaio/daytona/pkg/provider"
)

type TargetOptions struct {
	Region          string `json:"Region"`
	ImageId         string `json:"Image Id"`
	InstanceType    string `json:"Instance Type"`
	DeviceName      string `json:"Device Name"`
	VolumeSize      int    `json:"Volume Size"`
	VolumeType      string `json:"Volume Type"`
	AccessKeyId     string `json:"Access Key Id"`
	SecretAccessKey string `json:"Secret Access Key"`
}

func GetTargetManifest() *provider.ProviderTargetManifest {
	return &provider.ProviderTargetManifest{
		"Region": provider.ProviderTargetProperty{
			Type:         provider.ProviderTargetPropertyTypeString,
			DefaultValue: "us-east-1",
			Description:  "The geographic area where AWS resourced are hosted. Default is us-east-1.",
		},
		"Image Id": provider.ProviderTargetProperty{
			Type:         provider.ProviderTargetPropertyTypeString,
			DefaultValue: "ami-04a81a99f5ec58529",
			Description: "The ID of the Amazon Machine Image (AMI) to launch an instance. Default is ami-04a81a99f5ec58529. " +
				"How to find AMI that meets your needs  https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/finding-an-ami.html",
		},
		"Instance Type": provider.ProviderTargetProperty{
			Type:         provider.ProviderTargetPropertyTypeString,
			DefaultValue: "t2.micro",
			Description: "The type of instance to launch. Default is t2.micro. List of available instance types " +
				"https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/instance-types.html#AvailableInstanceTypes",
		},
		"Device Name": provider.ProviderTargetProperty{
			Type:         provider.ProviderTargetPropertyTypeString,
			DefaultValue: "/dev/sda1",
			Description:  "The device name for the volume. This is typically the root device name for specified AMI.",
		},
		"Volume Size": provider.ProviderTargetProperty{
			Type:         provider.ProviderTargetPropertyTypeInt,
			DefaultValue: "20",
			Description:  "The size of the instance volume, in GB. Default is 20 GB. It is recommended that the disk size should be more than 20 GB.",
		},
		"Volume Type": provider.ProviderTargetProperty{
			Type:         provider.ProviderTargetPropertyTypeString,
			DefaultValue: "gp3",
			Description: "The type of volume. Default is gp3. List of available volume types " +
				"https://docs.aws.amazon.com/ebs/latest/userguide/ebs-volume-types.html",
		},
		"Access Key Id": provider.ProviderTargetProperty{
			Type:        provider.ProviderTargetPropertyTypeString,
			InputMasked: true,
			Description: "If empty, it will be fetched from the AWS_ACCESS_KEY_ID environment variable.",
		},
		"Secret Access Key": provider.ProviderTargetProperty{
			Type:        provider.ProviderTargetPropertyTypeString,
			InputMasked: true,
			Description: "If empty, it will be fetched from the AWS_SECRET_ACCESS_KEY environment variable.",
		},
	}
}

// ParseTargetOptions parses the target options from the JSON string.
func ParseTargetOptions(optionsJson string) (*TargetOptions, error) {
	var targetOptions TargetOptions
	err := json.Unmarshal([]byte(optionsJson), &targetOptions)
	if err != nil {
		return nil, err
	}

	if targetOptions.AccessKeyId == "" {
		accessKeyId, ok := os.LookupEnv("AWS_ACCESS_KEY_ID")
		if ok {
			targetOptions.AccessKeyId = accessKeyId
		}
	}

	if targetOptions.SecretAccessKey == "" {
		secretAccessKey, ok := os.LookupEnv("AWS_SECRET_ACCESS_KEY")
		if ok {
			targetOptions.SecretAccessKey = secretAccessKey
		}
	}

	if targetOptions.AccessKeyId == "" {
		return nil, fmt.Errorf("access key id not set in env/target options")
	}

	if targetOptions.SecretAccessKey == "" {
		return nil, fmt.Errorf("secret access key not set in env/target options")
	}

	return &targetOptions, nil
}
