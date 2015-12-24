package config

import (
	"fmt"
	"gopkg.in/gcfg.v1"
	"os"
	"os/user"
	"path"
)

type ShellConfig struct {
	Rcfile    string
	No_Search bool
}

type SshCommonConfig struct {
	User               string
	Proj_Path          string
	Forward_Agent      bool
	Ignore_Known_Hosts bool
}

type SshHostConfig struct {
	SshCommonConfig
	Hostname string
}

type Ec2HostConfig struct {
	SshCommonConfig
	Access_Key string
	Secret_Key string
	Region     string
	Name       string
}

type HostConfig struct {
	Shell ShellConfig
	Ssh   map[string]*SshHostConfig
	Ec2   map[string]*Ec2HostConfig
}

var userHomedir = func() (string, error) {
	usr, err := user.Current()
	if err != nil {
		return "", fmt.Errorf("Determining home directory: %s", err)
	}
	return usr.HomeDir, nil
}

func LoadHostConfig() (*HostConfig, error) {
	homedir, err := userHomedir()
	if err != nil {
		return nil, err
	}
	filename := path.Join(homedir, ".proj.cfg")

	var config HostConfig

	// read the config, if it exists
	_, err = os.Stat(filename)
	if err == nil {
		err = gcfg.ReadFileInto(&config, filename)
		if err != nil {
			return nil, fmt.Errorf("While reading %q: %s", filename, err)
		}
	}

	return &config, nil
}
