package provider

import (
	"encoding/json"
	"os"
	"testing"
	"time"

	awsutil "github.com/daytonaio/daytona-provider-aws/pkg/provider/util"
	"github.com/daytonaio/daytona-provider-aws/pkg/types"
	"github.com/daytonaio/daytona/pkg/models"
	"github.com/daytonaio/daytona/pkg/provider"
)

var (
	accessKey = os.Getenv("AWS_ACCESS_KEY_ID")
	secretKey = os.Getenv("AWS_SECRET_ACCESS_KEY")
	region    = os.Getenv("AWS_DEFAULT_REGION")

	awsProvider   = &AWSProvider{}
	targetOptions = &types.TargetOptions{
		Region:          region,
		ImageId:         "ami-04a81a99f5ec58529",
		InstanceType:    "t2.micro",
		DeviceName:      "/dev/sda1",
		VolumeSize:      10,
		VolumeType:      "gp3",
		AccessKeyId:     accessKey,
		SecretAccessKey: secretKey,
	}

	targetReq *provider.TargetRequest
)

func TestCreateTarget(t *testing.T) {
	_, _ = awsProvider.CreateTarget(targetReq)

	_, err := awsutil.GetInstance(targetReq.Target, targetOptions)
	if err != nil {
		t.Fatalf("Error getting machine: %s", err)
	}
}

func TestGetTargetProviderMetadata(t *testing.T) {
	targetInfo, err := awsProvider.GetTargetProviderMetadata(targetReq)
	if err != nil {
		t.Fatalf("Error getting target info: %s", err)
	}

	var targetMetadata types.TargetMetadata
	err = json.Unmarshal([]byte(targetInfo), &targetMetadata)
	if err != nil {
		t.Fatalf("Error unmarshalling target metadata: %s", err)
	}

	instance, err := awsutil.GetInstance(targetReq.Target, targetOptions)
	if err != nil {
		t.Fatalf("Error getting machine: %s", err)
	}

	if targetMetadata.InstanceId != *instance.InstanceId {
		t.Fatalf("Expected instance id %s, got %s", targetMetadata.InstanceId, *instance.InstanceId)
	}
}

func TestDestroyTarget(t *testing.T) {
	_, err := awsProvider.DestroyTarget(targetReq)
	if err != nil {
		t.Fatalf("Error destroying target: %s", err)
	}
	time.Sleep(3 * time.Second)

	_, err = awsutil.GetInstance(targetReq.Target, targetOptions)
	if err == nil {
		t.Fatalf("Error destroyed target still exists")
	}
}

func init() {
	_, err := awsProvider.Initialize(provider.InitializeProviderRequest{
		BasePath:           "/tmp/targets",
		DaytonaDownloadUrl: "https://download.daytona.io/daytona/install.sh",
		DaytonaVersion:     "latest",
		ServerUrl:          "",
		ApiUrl:             "",
		WorkspaceLogsDir:   "/tmp/workspace/logs",
		TargetLogsDir:      "/tmp/target/logs",
	})
	if err != nil {
		panic(err)
	}

	opts, err := json.Marshal(targetOptions)
	if err != nil {
		panic(err)
	}

	targetReq = &provider.TargetRequest{
		Target: &models.Target{
			Id:   "123",
			Name: "target",
			TargetConfig: models.TargetConfig{
				Name: "test",
				ProviderInfo: models.ProviderInfo{
					Name:    "aws-provider",
					Version: "test",
				},
				Options: string(opts),
				Deleted: false,
			},
		},
	}

}
