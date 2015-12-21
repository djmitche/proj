package config

import (
	"fmt"
	"gopkg.in/gcfg.v1"
	"os/user"
	"path"
)

type SshCommonConfig struct {
	User          string
	Proj_Path     string
	Forward_Agent bool
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
	Ssh map[string]*SshHostConfig
	Ec2 map[string]*Ec2HostConfig
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
	err = gcfg.ReadFileInto(&config, filename)
	if err != nil {
		return nil, fmt.Errorf("While reading %q: %s", filename, err)
	}

	return &config, nil
}
