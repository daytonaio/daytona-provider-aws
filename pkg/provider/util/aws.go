package util

import (
	"encoding/base64"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/daytonaio/daytona-provider-aws/pkg/types"
	"github.com/daytonaio/daytona/pkg/models"
)

func CreateTarget(target *models.Target, opts *types.TargetOptions, initScript string) error {
	client, err := getEC2Client(opts)
	if err != nil {
		return err
	}

	envVars := target.EnvVars
	envVars["DAYTONA_AGENT_LOG_FILE_PATH"] = "/home/daytona/.daytona-agent.log"

	userData := `#!/bin/bash
useradd -m -d /home/daytona daytona

curl -fsSL https://get.docker.com | bash

# Modify Docker daemon configuration
cat > /etc/docker/daemon.json <<EOF
{
  "hosts": ["unix:///var/run/docker.sock", "tcp://0.0.0.0:2375"]
}
EOF

# Create a systemd drop-in file to modify the Docker service
mkdir -p /etc/systemd/system/docker.service.d
cat > /etc/systemd/system/docker.service.d/override.conf <<EOF
[Service]
ExecStart=
ExecStart=/usr/bin/dockerd
EOF

systemctl daemon-reload
systemctl restart docker
systemctl start docker

usermod -aG docker daytona

if grep -q sudo /etc/group; then
	usermod -aG sudo,docker daytona
elif grep -q wheel /etc/group; then
	usermod -aG wheel,docker daytona
fi

echo "daytona ALL=(ALL) NOPASSWD:ALL" > /etc/sudoers.d/91-daytona

`

	for k, v := range envVars {
		userData += fmt.Sprintf("export %s=%s\n", k, v)
	}
	userData += initScript
	userData += `
echo '[Unit]
Description=Daytona Agent Service
After=network.target

[Service]
User=daytona
ExecStart=/usr/local/bin/daytona agent --target
Restart=always
`

	for k, v := range envVars {
		userData += fmt.Sprintf("Environment='%s=%s'\n", k, v)
	}

	userData += `
[Install]
WantedBy=multi-user.target' > /etc/systemd/system/daytona-agent.service
systemctl daemon-reload
systemctl enable daytona-agent.service
systemctl start daytona-agent.service
`

	result, err := client.RunInstances(&ec2.RunInstancesInput{
		ImageId:      aws.String(opts.ImageId),
		InstanceType: aws.String(opts.InstanceType),
		MinCount:     aws.Int64(1),
		MaxCount:     aws.Int64(1),
		UserData:     aws.String(base64.StdEncoding.EncodeToString([]byte(userData))),
		BlockDeviceMappings: []*ec2.BlockDeviceMapping{
			{
				DeviceName: aws.String(opts.DeviceName),
				Ebs: &ec2.EbsBlockDevice{
					VolumeSize:          aws.Int64(int64(opts.VolumeSize)),
					VolumeType:          aws.String(opts.VolumeType),
					DeleteOnTermination: aws.Bool(true),
				},
			},
		},
		TagSpecifications: []*ec2.TagSpecification{
			{
				ResourceType: aws.String("instance"),
				Tags: []*ec2.Tag{
					{
						Key:   aws.String("Name"),
						Value: aws.String(fmt.Sprintf("daytona-%s", target.Id)),
					},
					{
						Key:   aws.String("WorkspaceID"),
						Value: aws.String(target.Id),
					},
				},
			},
		},
	})
	if err != nil {
		return err
	}

	return client.WaitUntilInstanceRunning(&ec2.DescribeInstancesInput{
		InstanceIds: []*string{result.Instances[0].InstanceId},
	})
}

func StartTarget(target *models.Target, opts *types.TargetOptions) error {
	client, err := getEC2Client(opts)
	if err != nil {
		return err
	}

	instance, err := getInstanceByWorkspaceID(client, target.Id)
	if err != nil {
		return err
	}

	if instance.State.String() == "running" {
		return nil
	}

	_, err = client.StartInstances(&ec2.StartInstancesInput{
		InstanceIds: []*string{instance.InstanceId},
	})
	if err != nil {
		return err
	}

	return client.WaitUntilInstanceRunning(&ec2.DescribeInstancesInput{
		InstanceIds: []*string{instance.InstanceId},
	})
}

func StopTarget(target *models.Target, opts *types.TargetOptions) error {
	client, err := getEC2Client(opts)
	if err != nil {
		return err
	}

	instance, err := getInstanceByWorkspaceID(client, target.Id)
	if err != nil {
		return err
	}

	_, err = client.StopInstances(&ec2.StopInstancesInput{
		InstanceIds: []*string{instance.InstanceId},
	})
	if err != nil {
		return err
	}

	return client.WaitUntilInstanceStopped(&ec2.DescribeInstancesInput{
		InstanceIds: []*string{instance.InstanceId},
	})
}

func DeleteTarget(target *models.Target, opts *types.TargetOptions) error {
	client, err := getEC2Client(opts)
	if err != nil {
		return err
	}

	instance, err := getInstanceByWorkspaceID(client, target.Id)
	if err != nil {
		return err
	}

	_, err = client.TerminateInstances(&ec2.TerminateInstancesInput{
		InstanceIds: []*string{instance.InstanceId},
	})
	return err
}

func GetInstance(target *models.Target, opts *types.TargetOptions) (*ec2.Instance, error) {
	client, err := getEC2Client(opts)
	if err != nil {
		return nil, err
	}

	return getInstanceByWorkspaceID(client, target.Id)
}

// getEC2Client  creates a new EC2 client using the provided target options.
func getEC2Client(opts *types.TargetOptions) (*ec2.EC2, error) {
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(opts.Region),
		Credentials: credentials.NewStaticCredentials(
			opts.AccessKeyId,
			opts.SecretAccessKey,
			"",
		),
	})
	if err != nil {
		return nil, err
	}

	return ec2.New(sess), nil
}

// getInstanceByWorkspaceID retrieves the first running or stopped EC2 instance
// associated with a given workspace ID.
func getInstanceByWorkspaceID(svc *ec2.EC2, workspaceID string) (*ec2.Instance, error) {
	result, err := svc.DescribeInstances(&ec2.DescribeInstancesInput{
		Filters: []*ec2.Filter{
			{
				Name:   aws.String("tag:WorkspaceID"),
				Values: []*string{aws.String(workspaceID)},
			},
			{
				Name:   aws.String("instance-state-name"),
				Values: []*string{aws.String("running"), aws.String("stopped")},
			},
		},
	})
	if err != nil {
		return nil, err
	}

	var instances []*ec2.Instance
	for _, reservation := range result.Reservations {
		instances = append(instances, reservation.Instances...)
	}

	return instances[0], nil
}
