package provider

import (
	"encoding/json"
	"os"
	"testing"
	"time"

	awsutil "github.com/daytonaio/daytona-provider-aws/pkg/provider/util"
	"github.com/daytonaio/daytona-provider-aws/pkg/types"
	"github.com/daytonaio/daytona/pkg/provider"
	"github.com/daytonaio/daytona/pkg/workspace"
)

var (
	accessKey = os.Getenv("AWS_ACCESS_KEY_ID")
	secretKey = os.Getenv("AWS_SECRET_ACCESS_KEY")

	awsProvider   = &AWSProvider{}
	targetOptions = &types.TargetOptions{
		Region:          "us-east-1",
		ImageId:         "ami-04a81a99f5ec58529",
		InstanceType:    "t2.micro",
		DeviceName:      "/dev/sda1",
		VolumeSize:      10,
		VolumeType:      "gp3",
		AccessKeyId:     accessKey,
		SecretAccessKey: secretKey,
	}

	workspaceReq *provider.WorkspaceRequest
)

func TestCreateWorkspace(t *testing.T) {
	_, _ = awsProvider.CreateWorkspace(workspaceReq)

	_, err := awsutil.GetInstance(workspaceReq.Workspace, targetOptions)
	if err != nil {
		t.Fatalf("Error getting machine: %s", err)
	}
}

func TestWorkspaceInfo(t *testing.T) {
	workspaceInfo, err := awsProvider.GetWorkspaceInfo(workspaceReq)
	if err != nil {
		t.Fatalf("Error getting workspace info: %s", err)
	}

	var workspaceMetadata types.WorkspaceMetadata
	err = json.Unmarshal([]byte(workspaceInfo.ProviderMetadata), &workspaceMetadata)
	if err != nil {
		t.Fatalf("Error unmarshalling workspace metadata: %s", err)
	}

	instance, err := awsutil.GetInstance(workspaceReq.Workspace, targetOptions)
	if err != nil {
		t.Fatalf("Error getting machine: %s", err)
	}

	if workspaceMetadata.InstanceId != *instance.InstanceId {
		t.Fatalf("Expected instance id %s, got %s", workspaceMetadata.InstanceId, *instance.InstanceId)
	}
}

func TestDestroyWorkspace(t *testing.T) {
	_, err := awsProvider.DestroyWorkspace(workspaceReq)
	if err != nil {
		t.Fatalf("Error destroying workspace: %s", err)
	}
	time.Sleep(3 * time.Second)

	_, err = awsutil.GetInstance(workspaceReq.Workspace, targetOptions)
	if err == nil {
		t.Fatalf("Error destroyed workspace still exists")
	}
}

func init() {
	_, err := awsProvider.Initialize(provider.InitializeProviderRequest{
		BasePath:           "/tmp/workspaces",
		DaytonaDownloadUrl: "https://download.daytona.io/daytona/install.sh",
		DaytonaVersion:     "latest",
		ServerUrl:          "",
		ApiUrl:             "",
		LogsDir:            "/tmp/logs",
	})
	if err != nil {
		panic(err)
	}

	opts, err := json.Marshal(targetOptions)
	if err != nil {
		panic(err)
	}

	workspaceReq = &provider.WorkspaceRequest{
		TargetOptions: string(opts),
		Workspace: &workspace.Workspace{
			Id:   "123",
			Name: "workspace",
		},
	}

}
