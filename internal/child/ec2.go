package child

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/djmitche/proj/internal/config"
	"github.com/djmitche/proj/internal/ssh"
	"log"
	"net"
	"time"
)

func setupEc2(ec2HostConfig *config.Ec2HostConfig, childConfig *config.ChildConfig) (*ec2.EC2, error) {
	var creds *credentials.Credentials

	if ec2HostConfig.Access_Key != "" {
		if ec2HostConfig.Secret_Key == "" {
			return nil, fmt.Errorf("config includes ec2 id but not secret")
		}
		creds = credentials.NewStaticCredentials(
			ec2HostConfig.Access_Key,
			ec2HostConfig.Secret_Key,
			"")
	}

	svc := ec2.New(session.New(), &aws.Config{
		Region:      aws.String(ec2HostConfig.Region),
		Credentials: creds})

	log.Printf("Connected to EC2 in region %s", ec2HostConfig.Region)
	return svc, nil
}

func findInstance(ec2HostConfig *config.Ec2HostConfig, svc *ec2.EC2) (*ec2.Instance, error) {
	region, name := ec2HostConfig.Region, ec2HostConfig.Name

	log.Printf("Searching for instance named %q in region %s", name, region)

	params := &ec2.DescribeInstancesInput{
		Filters: []*ec2.Filter{
			{
				Name: aws.String("tag:Name"),
				Values: []*string{
					aws.String(ec2HostConfig.Name),
				},
			},
		},
	}
	resp, err := svc.DescribeInstances(params)
	if err != nil {
		return nil, err
	}

	if len(resp.Reservations) < 1 {
		return nil, fmt.Errorf("no instance found in %s with name %s", region, name)
	}
	if len(resp.Reservations) > 1 || len(resp.Reservations[0].Instances) != 1 {
		return nil, fmt.Errorf("multiple instances found in %s with name %s", region, name)
	}
	return resp.Reservations[0].Instances[0], nil
}

func startInstance(ec2HostConfig *config.Ec2HostConfig, instance *ec2.Instance, svc *ec2.EC2) error {
	startCalled := false
	instanceId := *instance.InstanceId

	// wait until the state is anything but "pending"
statePoll:
	for {
		state, err := getInstanceState(instanceId, svc)
		if err != nil {
			return err
		}

		log.Printf("Instance %s is in state %s", instanceId, state)
		switch state {
		case "pending", "shutting-down", "stopping":
			// for any of the transient states, just wait
			time.Sleep(time.Second / 2)

		case "running":
			// running is the state we want to get to
			break statePoll

		case "stopped":
			log.Printf("starting instance %s", instanceId)
			if startCalled {
				return fmt.Errorf("instance %s did not enter the running state", instanceId)
			}
			// start a stopped instance
			params := &ec2.StartInstancesInput{
				InstanceIds: []*string{aws.String(instanceId)},
			}
			_, err := svc.StartInstances(params)
			if err != nil {
				return err
			}
			startCalled = true

		case "terminated":
			return fmt.Errorf("Instance is terminated")
		}
	}

	// wait for SSH port to be open, too
	for {
		// re-search for the instance, since it probably didn't have a public ip address
		// when it was disconnected -- TODO refactor
		if instance.PublicIpAddress == nil {
			var err error
			instance, err = findInstance(ec2HostConfig, svc)
			if err != nil {
				return fmt.Errorf("while searcihng for running instance: %s", err)
			}
		}
		conn, err := net.Dial("tcp", fmt.Sprintf("%s:22", *instance.PublicIpAddress))
		if err != nil {
			log.Printf("connecting to port 22: %s; retrying", err)
			time.Sleep(time.Second / 2)
		} else {
			conn.Close()
			break
		}
	}

	return nil
}

// get the current state for an EC2 instance
func getInstanceState(instanceId string, svc *ec2.EC2) (string, error) {
	params := &ec2.DescribeInstancesInput{
		InstanceIds: []*string{&instanceId},
	}
	resp, err := svc.DescribeInstances(params)
	if err != nil {
		return "", err
	}

	if len(resp.Reservations) < 1 {
		return "", fmt.Errorf("instance %s not found", instanceId)
	}
	if len(resp.Reservations) > 1 || len(resp.Reservations[0].Instances) != 1 {
		return "", fmt.Errorf("Multiple instances %s found (?!)", instanceId)
	}
	return *resp.Reservations[0].Instances[0].State.Name, nil
}

func ec2Child(info *childInfo) error {
	instanceName := info.childConfig.Ec2.Instance
	ec2HostConfig, ok := info.hostConfig.Ec2[instanceName]
	if !ok {
		return fmt.Errorf("EC2 instance %q not defined in configuration file", instanceName)
	}

	svc, err := setupEc2(ec2HostConfig, info.childConfig)
	if err != nil {
		return fmt.Errorf("while setting up ec2 access: %s", err)
	}

	// look up the instance matching this description
	instance, err := findInstance(ec2HostConfig, svc)
	if err != nil {
		return fmt.Errorf("while searching for ec2 instance: %s", err)
	}
	log.Printf("Found instance id %s (type %s)", *instance.InstanceId, *instance.InstanceType)

	err = startInstance(ec2HostConfig, instance, svc)
	if err != nil {
		return fmt.Errorf("while starting instance: %s", err)
	}

	// re-fetch the instance to get an IP address
	if instance.PublicIpAddress == nil {
		instance, err = findInstance(ec2HostConfig, svc)
		if err != nil {
			return fmt.Errorf("while searcihng for running instance: %s", err)
		}
	}

	return ssh.Run(*instance.PublicIpAddress, &ec2HostConfig.SshCommonConfig, info.path)
}

func init() {
	childFuncs["ec2"] = ec2Child
}
