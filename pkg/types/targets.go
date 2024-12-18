package types

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/daytonaio/daytona/pkg/models"
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

func GetTargetConfigManifest() *models.TargetConfigManifest {
	return &models.TargetConfigManifest{
		"Region": models.TargetConfigProperty{
			Type:         models.TargetConfigPropertyTypeString,
			DefaultValue: "us-east-1",
			Description: "The geographic area where AWS resourced are hosted. List of available regions:\n" +
				"https://docs.aws.amazon.com/general/latest/gr/rande.html\n" +
				"Leave blank if you've set the AWS_DEFAULT_REGION environment variable, or enter your region here.",
		},
		"Instance Type": models.TargetConfigProperty{
			Type:         models.TargetConfigPropertyTypeString,
			DefaultValue: "t2.micro",
			Description: "The type of instance to launch. Default is t2.micro.\nList of available instance types:\n" +
				"https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/instance-types.html#AvailableInstanceTypes",
		},
		"Image Id": models.TargetConfigProperty{
			Type:         models.TargetConfigPropertyTypeString,
			DefaultValue: "ami-04a81a99f5ec58529",
			Description: "The ID of the Amazon Machine Image (AMI) to launch an instance. Default is ami-04a81a99f5ec58529.\n" +
				"How to find AMI that meets your needs:\nhttps://docs.aws.amazon.com/AWSEC2/latest/UserGuide/finding-an-ami.html",
		},
		"Device Name": models.TargetConfigProperty{
			Type:         models.TargetConfigPropertyTypeString,
			DefaultValue: "/dev/sda1",
			Description: "The device name for the volume. This is typically the root device name for specified AMI.\n" +
				"List of device names:\nhttps://docs.aws.amazon.com/AWSEC2/latest/UserGuide/device_naming.html",
		},
		"Volume Size": models.TargetConfigProperty{
			Type:         models.TargetConfigPropertyTypeInt,
			DefaultValue: "20",
			Description: "The size of the instance volume, in GB. Default is 20 GB.It is recommended that the disk size should be more than 20 GB.\n" +
				"List of volume size limits:\nhttps://docs.aws.amazon.com/AWSEC2/latest/UserGuide/volume_limits.html",
		},
		"Volume Type": models.TargetConfigProperty{
			Type:         models.TargetConfigPropertyTypeString,
			DefaultValue: "gp3",
			Description: "The type of volume. Default is gp3.\n" +
				"List of available volume types:\nhttps://docs.aws.amazon.com/ebs/latest/userguide/ebs-volume-types.html",
		},
		"Access Key Id": models.TargetConfigProperty{
			Type:        models.TargetConfigPropertyTypeString,
			InputMasked: true,
			Description: "Find this in the AWS Console under \"My Security Credentials\"\nhttps://aws.amazon.com/premiumsupport/knowledge-center/manage-access-keys/\n" +
				"Leave blank if you've set the AWS_ACCESS_KEY_ID environment variable, or enter your Id here.",
		},
		"Secret Access Key": models.TargetConfigProperty{
			Type:        models.TargetConfigPropertyTypeString,
			InputMasked: true,
			Description: "Find this in the AWS Console under \"My Security Credentials\"\nhttps://aws.amazon.com/premiumsupport/knowledge-center/manage-access-keys/\n" +
				"Leave blank if you've set the AWS_SECRET_ACCESS_KEY environment variable, or enter your key here.",
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

	if targetOptions.Region == "" {
		region, ok := os.LookupEnv("AWS_DEFAULT_REGION")
		if ok {
			targetOptions.Region = region
		}
	}

	if targetOptions.AccessKeyId == "" {
		return nil, fmt.Errorf("access key id not set in env/target options")
	}

	if targetOptions.SecretAccessKey == "" {
		return nil, fmt.Errorf("secret access key not set in env/target options")
	}

	if targetOptions.Region == "" {
		return nil, fmt.Errorf("region not set in env/target options")
	}

	return &targetOptions, nil
}
