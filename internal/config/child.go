package config

import (
	"fmt"
	"gopkg.in/gcfg.v1"
)

type ChildCommonConfig struct {
	Prepend string
	//Shellrc string -- TODO
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
	Shell struct {
		ChildCommonConfig
		Command string
	}
}

func (cc *ChildConfig) Common() *ChildCommonConfig {
	switch cc.Type {
	case "cd":
		return &cc.Cd.ChildCommonConfig
	case "ssh":
		return &cc.Ssh.ChildCommonConfig
	case "shell":
		return &cc.Shell.ChildCommonConfig
	case "ec2":
		return &cc.Ec2.ChildCommonConfig
	default:
		panic("Invalid child config type")
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
	} else if config.Ssh.Host != "" {
		config.Type = "ssh"
	} else if config.Shell.Command != "" {
		config.Type = "shell"
	} else if config.Ec2.Instance != "" {
		config.Type = "ec2"
	} else {
		return nil, fmt.Errorf("Unknown child type")
	}

	return &config, nil
}
