package config

import (
	"fmt"
	"gopkg.in/gcfg.v1"
)

type ChildCommonConfig struct {
	Prepend string
	Shellrc string
}

type ChildConfig struct {
	Type string

	Cd struct {
		ChildCommonConfig
		Dir string
	}
	Ssh struct {
		ChildCommonConfig
		Host string
	}
	Ec2 struct {
		ChildCommonConfig
		Instance string
	}
}

func LoadChildConfig(filename string) (*ChildConfig, error) {
	var config ChildConfig
	err := gcfg.ReadFileInto(&config, filename)
	if err != nil {
		return nil, fmt.Errorf("While reading %q: %s", filename, err)
	}

	// Figure out the type.  Note that if the user includes multiple sections,
	// this will just use the last one in the list.  Note that this depends on
	// each type having a mandatory, distinct key.
	if config.Cd.Dir != "" {
		config.Type = "cd"
	}
	if config.Ssh.Host != "" {
		config.Type = "ssh"
	}
	if config.Ec2.Instance != "" {
		config.Type = "ec2"
	}

	return &config, nil
}
