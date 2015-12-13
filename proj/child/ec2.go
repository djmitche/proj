package child

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/djmitche/proj/proj/ssh"
	"log"
	"net"
	"time"
)

type ec2Config struct {
	// ec2 API access
	id     string
	secret string

	// instance identifier
	region string
	name   string

	// access information
	user     string
	config   string
	projPath string
}

func setupEc2(cfg *ec2Config) (*ec2.EC2, error) {
	var creds *credentials.Credentials

	if cfg.id != "" {
		if cfg.secret == "" {
			return nil, fmt.Errorf("config includes ec2 id but not secret")
		}
		creds = credentials.NewStaticCredentials(cfg.id, cfg.secret, "")
	}

	svc := ec2.New(session.New(), &aws.Config{
		Region:      aws.String(cfg.region),
		Credentials: creds})

	log.Printf("Connected to EC2 in region %s", cfg.region)
	return svc, nil
}

func findInstance(cfg *ec2Config, svc *ec2.EC2) (*ec2.Instance, error) {
	log.Printf("Searching for instance named %q in region %s", cfg.name, cfg.region)

	params := &ec2.DescribeInstancesInput{
		Filters: []*ec2.Filter{
			{
				Name: aws.String("tag:Name"),
				Values: []*string{
					aws.String(cfg.name),
				},
			},
		},
	}
	resp, err := svc.DescribeInstances(params)
	if err != nil {
		return nil, err
	}

	if len(resp.Reservations) < 1 {
		return nil, fmt.Errorf("no instance found in %s with name %s", cfg.region, cfg.name)
	}
	if len(resp.Reservations) > 1 || len(resp.Reservations[0].Instances) != 1 {
		return nil, fmt.Errorf("multiple instances found in %s with name %s", cfg.region, cfg.name)
	}
	return resp.Reservations[0].Instances[0], nil
}

func startInstance(cfg *ec2Config, instance *ec2.Instance, svc *ec2.EC2) error {
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
			instance, err = findInstance(cfg, svc)
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
	var cfg ec2Config

	argsMap, ok := info.args.(map[interface{}]interface{})
	if !ok {
		return fmt.Errorf("ec2 child must at least have keys 'region' and 'name'")
	}

	var cfgFields = []struct {
		name string
		val  *string
	}{
		{"id", &cfg.id},
		{"secret", &cfg.secret},
		{"region", &cfg.region},
		{"name", &cfg.name},
		{"user", &cfg.user},
		{"config", &cfg.config},
		{"proj", &cfg.projPath},
	}

	for _, fld := range cfgFields {
		arg, ok := argsMap[fld.name]
		if ok {
			argStr, ok := arg.(string)
			if ok {
				*fld.val = argStr
			} else {
				return fmt.Errorf("%s should be a string; got %q", fld.name, arg)
			}
		}
	}

	if cfg.region == "" || cfg.name == "" {
		return fmt.Errorf("ec2 child must at least have keys 'region' and 'name'")
	}

	svc, err := setupEc2(&cfg)
	if err != nil {
		return fmt.Errorf("while setting up ec2 access: %s", err)
	}

	// look up the instance matching this description
	instance, err := findInstance(&cfg, svc)
	if err != nil {
		return fmt.Errorf("while searching for ec2 instance: %s", err)
	}
	log.Printf("Found instance id %s (type %s)", *instance.InstanceId, *instance.InstanceType)

	err = startInstance(&cfg, instance, svc)
	if err != nil {
		return fmt.Errorf("while starting instance: %s", err)
	}

	// re-fetch the instance to get an IP address
	if instance.PublicIpAddress == nil {
		instance, err = findInstance(&cfg, svc)
		if err != nil {
			return fmt.Errorf("while searcihng for running instance: %s", err)
		}
	}

	return ssh.Run(&ssh.Config{
		User:           cfg.user,
		Host:           *instance.PublicIpAddress,
		ConfigFilename: cfg.config,
		ProjPath:       cfg.projPath,
		Path:           info.path,
	})
}

func init() {
	childFuncs["ec2"] = ec2Child
}
