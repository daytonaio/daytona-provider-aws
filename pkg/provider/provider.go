package provider

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"path"
	"time"

	"github.com/daytonaio/daytona-provider-aws/internal"
	logwriters "github.com/daytonaio/daytona-provider-aws/internal/log"
	awsutil "github.com/daytonaio/daytona-provider-aws/pkg/provider/util"
	"github.com/daytonaio/daytona-provider-aws/pkg/types"
	"github.com/daytonaio/daytona/pkg/agent/ssh/config"
	"github.com/daytonaio/daytona/pkg/docker"
	"github.com/daytonaio/daytona/pkg/models"
	"github.com/daytonaio/daytona/pkg/ssh"
	"github.com/daytonaio/daytona/pkg/tailscale"
	"tailscale.com/tsnet"

	"github.com/daytonaio/daytona/pkg/logs"
	"github.com/daytonaio/daytona/pkg/provider"
	"github.com/daytonaio/daytona/pkg/provider/util"
)

type AWSProvider struct {
	BasePath           *string
	DaytonaDownloadUrl *string
	DaytonaVersion     *string
	ServerUrl          *string
	NetworkKey         *string
	ApiUrl             *string
	ApiKey             *string
	ApiPort            *uint32
	ServerPort         *uint32
	WorkspaceLogsDir   *string
	TargetLogsDir      *string
	tsnetConn          *tsnet.Server
}

func (a *AWSProvider) Initialize(req provider.InitializeProviderRequest) (*util.Empty, error) {
	a.BasePath = &req.BasePath
	a.DaytonaDownloadUrl = &req.DaytonaDownloadUrl
	a.DaytonaVersion = &req.DaytonaVersion
	a.ServerUrl = &req.ServerUrl
	a.NetworkKey = &req.NetworkKey
	a.ApiUrl = &req.ApiUrl
	a.ApiKey = req.ApiKey
	a.ApiPort = &req.ApiPort
	a.ServerPort = &req.ServerPort
	a.WorkspaceLogsDir = &req.WorkspaceLogsDir
	a.TargetLogsDir = &req.TargetLogsDir

	return new(util.Empty), nil
}

func (a *AWSProvider) GetInfo() (models.ProviderInfo, error) {
	label := "AWS"

	return models.ProviderInfo{
		Label:                &label,
		Name:                 "aws-provider",
		Version:              internal.Version,
		TargetConfigManifest: *types.GetTargetConfigManifest(),
	}, nil
}

func (a *AWSProvider) GetPresetTargetConfigs() (*[]provider.TargetConfig, error) {
	return new([]provider.TargetConfig), nil
}

func (a *AWSProvider) CreateTarget(targetReq *provider.TargetRequest) (*util.Empty, error) {
	if a.DaytonaDownloadUrl == nil {
		return nil, errors.New("DaytonaDownloadUrl not set. Did you forget to call Initialize")
	}
	logWriter, cleanupFunc := a.getTargetLogWriter(targetReq.Target.Id, targetReq.Target.Name)
	defer cleanupFunc()

	targetOptions, err := types.ParseTargetOptions(targetReq.Target.TargetConfig.Options)
	if err != nil {
		logWriter.Write([]byte("Failed to parse target options: " + err.Error() + "\n"))
		return nil, err
	}

	ec2spinner := logwriters.ShowSpinner(logWriter, "Creating EC2 instance", "EC2 instance created")
	initScript := fmt.Sprintf(`curl -sfL -H "Authorization: Bearer %s" %s | bash`, targetReq.Target.ApiKey, *a.DaytonaDownloadUrl)

	err = awsutil.CreateTarget(targetReq.Target, targetOptions, initScript)
	close(ec2spinner)
	if err != nil {
		logWriter.Write([]byte("Failed to create workspace: " + err.Error() + "\n"))
		return nil, err
	}

	agentSpinner := logwriters.ShowSpinner(logWriter, "Waiting for the agent to start", "Agent started")

	err = a.waitForDial(targetReq.Target.Id, 10*time.Minute)
	close(agentSpinner)
	if err != nil {
		logWriter.Write([]byte("Failed to dial: " + err.Error() + "\n"))
		return nil, err
	}

	client, err := a.getDockerClient(targetReq.Target.Id)
	if err != nil {
		logWriter.Write([]byte("Failed to get client: " + err.Error() + "\n"))
		return nil, err
	}

	targetId := getTargetDir(targetReq.Target.Id)
	sshClient, err := tailscale.NewSshClient(a.tsnetConn, &ssh.SessionConfig{
		Hostname: targetReq.Target.Id,
		Port:     config.SSH_PORT,
	})
	if err != nil {
		logWriter.Write([]byte("Failed to create ssh client: " + err.Error() + "\n"))
		return new(util.Empty), err
	}
	defer sshClient.Close()

	return new(util.Empty), client.CreateTarget(targetReq.Target, targetId, logWriter, sshClient)
}

func (a *AWSProvider) StartTarget(targetReq *provider.TargetRequest) (*util.Empty, error) {
	logWriter, cleanupFunc := a.getTargetLogWriter(targetReq.Target.Id, targetReq.Target.Name)
	defer cleanupFunc()

	targetOptions, err := types.ParseTargetOptions(targetReq.Target.TargetConfig.Options)
	if err != nil {
		logWriter.Write([]byte("Failed to parse target options: " + err.Error() + "\n"))
		return nil, err
	}

	err = a.waitForDial(targetReq.Target.Id, 10*time.Minute)
	if err != nil {
		logWriter.Write([]byte("Failed to dial: " + err.Error() + "\n"))
		return nil, err
	}

	err = awsutil.StartTarget(targetReq.Target, targetOptions)
	if err != nil {
		return nil, err
	}

	return new(util.Empty), nil
}

func (a *AWSProvider) StopTarget(targetReq *provider.TargetRequest) (*util.Empty, error) {
	logWriter, cleanupFunc := a.getTargetLogWriter(targetReq.Target.Id, targetReq.Target.Name)
	defer cleanupFunc()

	targetOptions, err := types.ParseTargetOptions(targetReq.Target.TargetConfig.Options)
	if err != nil {
		logWriter.Write([]byte("Failed to parse target options: " + err.Error() + "\n"))
		return nil, err
	}

	return new(util.Empty), awsutil.StopTarget(targetReq.Target, targetOptions)
}

func (a *AWSProvider) DestroyTarget(targetReq *provider.TargetRequest) (*util.Empty, error) {
	logWriter, cleanupFunc := a.getTargetLogWriter(targetReq.Target.Id, targetReq.Target.Name)
	defer cleanupFunc()

	targetOptions, err := types.ParseTargetOptions(targetReq.Target.TargetConfig.Options)
	if err != nil {
		logWriter.Write([]byte("Failed to parse target options: " + err.Error() + "\n"))
		return nil, err
	}

	return new(util.Empty), awsutil.DeleteTarget(targetReq.Target, targetOptions)
}

func (a *AWSProvider) GetTargetProviderMetadata(targetReq *provider.TargetRequest) (string, error) {
	logWriter, cleanupFunc := a.getTargetLogWriter(targetReq.Target.Id, targetReq.Target.Name)
	defer cleanupFunc()

	targetOptions, err := types.ParseTargetOptions(targetReq.Target.TargetConfig.Options)
	if err != nil {
		logWriter.Write([]byte("Failed to parse target options: " + err.Error() + "\n"))
		return "", err
	}

	instance, err := awsutil.GetInstance(targetReq.Target, targetOptions)
	if err != nil {
		logWriter.Write([]byte("Failed to get machine: " + err.Error() + "\n"))
		return "", err

	}

	tags := map[string]string{}
	for _, tag := range instance.Tags {
		tags[*tag.Key] = *tag.Value
	}

	metadata := types.TargetMetadata{
		InstanceId: *instance.InstanceId,
		IsRunning:  instance.State.String() == "running",
		Created:    instance.LaunchTime.String(),
		Tags:       tags,
	}
	jsonMetadata, err := json.Marshal(metadata)
	if err != nil {
		return "", err
	}

	return string(jsonMetadata), nil
}

func (a *AWSProvider) CheckRequirements() (*[]provider.RequirementStatus, error) {
	results := []provider.RequirementStatus{}
	return &results, nil
}

func (a *AWSProvider) CreateWorkspace(workspaceReq *provider.WorkspaceRequest) (*util.Empty, error) {
	logWriter, cleanupFunc := a.getWorkspaceLogWriter(workspaceReq.Workspace.Id, workspaceReq.Workspace.Name)
	defer cleanupFunc()
	logWriter.Write([]byte("\033[?25h\n"))

	dockerClient, err := a.getDockerClient(workspaceReq.Workspace.Target.Id)
	if err != nil {
		logWriter.Write([]byte("Failed to get docker client: " + err.Error() + "\n"))
		return nil, err
	}

	sshClient, err := tailscale.NewSshClient(a.tsnetConn, &ssh.SessionConfig{
		Hostname: workspaceReq.Workspace.Target.Id,
		Port:     config.SSH_PORT,
	})
	if err != nil {
		logWriter.Write([]byte("Failed to create ssh client: " + err.Error() + "\n"))
		return new(util.Empty), err
	}
	defer sshClient.Close()

	return new(util.Empty), dockerClient.CreateWorkspace(&docker.CreateWorkspaceOptions{
		Workspace:           workspaceReq.Workspace,
		WorkspaceDir:        getWorkspaceDir(workspaceReq),
		ContainerRegistries: workspaceReq.ContainerRegistries,
		BuilderImage:        workspaceReq.BuilderImage,
		LogWriter:           logWriter,
		Gpc:                 workspaceReq.GitProviderConfig,
		SshClient:           sshClient,
	})
}

func (a *AWSProvider) StartWorkspace(workspaceReq *provider.WorkspaceRequest) (*util.Empty, error) {
	if a.DaytonaDownloadUrl == nil {
		return nil, errors.New("DaytonaDownloadUrl not set. Did you forget to call Initialize")
	}
	logWriter, cleanupFunc := a.getWorkspaceLogWriter(workspaceReq.Workspace.Id, workspaceReq.Workspace.Name)
	defer cleanupFunc()

	dockerClient, err := a.getDockerClient(workspaceReq.Workspace.Target.Id)
	if err != nil {
		logWriter.Write([]byte("Failed to get docker client: " + err.Error() + "\n"))
		return nil, err
	}

	sshClient, err := tailscale.NewSshClient(a.tsnetConn, &ssh.SessionConfig{
		Hostname: workspaceReq.Workspace.Target.Id,
		Port:     config.SSH_PORT,
	})
	if err != nil {
		logWriter.Write([]byte("Failed to create ssh client: " + err.Error() + "\n"))
		return new(util.Empty), err
	}
	defer sshClient.Close()

	return new(util.Empty), dockerClient.StartWorkspace(&docker.CreateWorkspaceOptions{
		Workspace:           workspaceReq.Workspace,
		WorkspaceDir:        getWorkspaceDir(workspaceReq),
		ContainerRegistries: workspaceReq.ContainerRegistries,
		BuilderImage:        workspaceReq.BuilderImage,
		LogWriter:           logWriter,
		Gpc:                 workspaceReq.GitProviderConfig,
		SshClient:           sshClient,
	}, *a.DaytonaDownloadUrl)
}

func (a *AWSProvider) StopWorkspace(workspaceReq *provider.WorkspaceRequest) (*util.Empty, error) {
	logWriter, cleanupFunc := a.getWorkspaceLogWriter(workspaceReq.Workspace.Id, workspaceReq.Workspace.Name)
	defer cleanupFunc()

	dockerClient, err := a.getDockerClient(workspaceReq.Workspace.Target.Id)
	if err != nil {
		logWriter.Write([]byte("Failed to get docker client: " + err.Error() + "\n"))
		return nil, err
	}

	return new(util.Empty), dockerClient.StopWorkspace(workspaceReq.Workspace, logWriter)
}

func (a *AWSProvider) DestroyWorkspace(workspaceReq *provider.WorkspaceRequest) (*util.Empty, error) {
	logWriter, cleanupFunc := a.getWorkspaceLogWriter(workspaceReq.Workspace.Id, workspaceReq.Workspace.Name)
	defer cleanupFunc()

	dockerClient, err := a.getDockerClient(workspaceReq.Workspace.Target.Id)
	if err != nil {
		logWriter.Write([]byte("Failed to get docker client: " + err.Error() + "\n"))
		return nil, err
	}

	sshClient, err := tailscale.NewSshClient(a.tsnetConn, &ssh.SessionConfig{
		Hostname: workspaceReq.Workspace.Target.Id,
		Port:     config.SSH_PORT,
	})
	if err != nil {
		logWriter.Write([]byte("Failed to create ssh client: " + err.Error() + "\n"))
		return new(util.Empty), err
	}
	defer sshClient.Close()

	return new(util.Empty), dockerClient.DestroyWorkspace(workspaceReq.Workspace, getWorkspaceDir(workspaceReq), sshClient)
}

func (a *AWSProvider) GetWorkspaceProviderMetadata(workspaceReq *provider.WorkspaceRequest) (string, error) {
	logWriter, cleanupFunc := a.getWorkspaceLogWriter(workspaceReq.Workspace.Id, workspaceReq.Workspace.Name)
	defer cleanupFunc()

	dockerClient, err := a.getDockerClient(workspaceReq.Workspace.Target.Id)
	if err != nil {
		logWriter.Write([]byte("Failed to get docker client: " + err.Error() + "\n"))
		return "", err
	}

	return dockerClient.GetWorkspaceProviderMetadata(workspaceReq.Workspace)
}

func (a *AWSProvider) getWorkspaceLogWriter(workspaceId, workspaceName string) (io.Writer, func()) {
	logWriter := io.MultiWriter(&logwriters.InfoLogWriter{})
	cleanupFunc := func() {}

	if a.WorkspaceLogsDir != nil {
		loggerFactory := logs.NewLoggerFactory(logs.LoggerFactoryConfig{
			LogsDir:     *a.WorkspaceLogsDir,
			ApiUrl:      a.ApiUrl,
			ApiKey:      a.ApiKey,
			ApiBasePath: &logs.ApiBasePathWorkspace,
		})
		workspaceLogWriter, err := loggerFactory.CreateLogger(workspaceId, workspaceName, logs.LogSourceProvider)
		if err == nil {
			logWriter = io.MultiWriter(&logwriters.InfoLogWriter{}, workspaceLogWriter)
			cleanupFunc = func() { workspaceLogWriter.Close() }
		}
	}

	return logWriter, cleanupFunc
}

func (a *AWSProvider) getTargetLogWriter(targetId, targetName string) (io.Writer, func()) {
	logWriter := io.MultiWriter(&logwriters.InfoLogWriter{})
	cleanupFunc := func() {}

	if a.TargetLogsDir != nil {
		loggerFactory := logs.NewLoggerFactory(logs.LoggerFactoryConfig{
			LogsDir:     *a.TargetLogsDir,
			ApiUrl:      a.ApiUrl,
			ApiKey:      a.ApiKey,
			ApiBasePath: &logs.ApiBasePathTarget,
		})
		workspaceLogWriter, err := loggerFactory.CreateLogger(targetId, targetName, logs.LogSourceProvider)
		if err == nil {
			logWriter = io.MultiWriter(&logwriters.InfoLogWriter{}, workspaceLogWriter)
			cleanupFunc = func() { workspaceLogWriter.Close() }
		}
	}

	return logWriter, cleanupFunc
}

func getWorkspaceDir(workspaceReq *provider.WorkspaceRequest) string {
	return path.Join(
		getTargetDir(workspaceReq.Workspace.TargetId),
		workspaceReq.Workspace.Id,
		workspaceReq.Workspace.WorkspaceFolderName(),
	)
}

func getTargetDir(targetId string) string {
	return fmt.Sprintf("/home/daytona/%s", targetId)
}
