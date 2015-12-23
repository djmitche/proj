package config

import (
	"io/ioutil"
	"os"
	"path"
	"testing"
)

func loadTestConfig(config string) (*HostConfig, error) {
	// make a temporary "homedir"
	tmpdir, err := ioutil.TempDir("", "host_test")
	if err != nil {
		return nil, err
	}
	defer func() { os.RemoveAll(tmpdir) }()

	// write out the contents
	filename := path.Join(tmpdir, ".proj.cfg")
	err = ioutil.WriteFile(filename, []byte(config), 0700)
	if err != nil {
		return nil, err
	}

	// patch in a new userHomedir function
	oldUserHomedir := userHomedir
	userHomedir = func() (string, error) { return tmpdir, nil }
	defer func() { userHomedir = oldUserHomedir }()

	// load the config
	return LoadHostConfig()
}

func TestEmptyConfig(t *testing.T) {
	config, err := loadTestConfig("")
	if err != nil {
		t.Log(err)
		t.FailNow()
	}

	if len(config.Ssh) != 0 {
		t.Errorf("nonzero number of SSH entries")
	}
	if len(config.Ec2) != 0 {
		t.Errorf("nonzero number of EC2 entries")
	}
}

const fullConfig = `
[ec2 "foo"]
user = test
proj-path = /with spaces/
forward-agent =  yes
ignore-known-hosts = yes
access-key = 123
secret-key = 456
region = "us-north-1"
name = "foo-server"

[ssh "bar"]
hostname = "bar.com"
user = test
`

func TestFullConfig(t *testing.T) {
	config, err := loadTestConfig(fullConfig)
	if err != nil {
		t.Log(err)
		t.FailNow()
	}

	if (*config.Ec2["foo"] != Ec2HostConfig{
		SshCommonConfig: SshCommonConfig{
			User:               "test",
			Proj_Path:          "/with spaces/",
			Forward_Agent:      true,
			Ignore_Known_Hosts: true,
		},
		Access_Key: "123",
		Secret_Key: "456",
		Region:     "us-north-1",
		Name:       "foo-server",
	}) {
		t.Errorf("got incorrect Ec2HostConfig %#v", config.Ec2["foo"])
	}

	if (*config.Ssh["bar"] != SshHostConfig{
		SshCommonConfig: SshCommonConfig{
			User:               "test",
			Proj_Path:          "",
			Forward_Agent:      false,
			Ignore_Known_Hosts: false,
		},
		Hostname: "bar.com",
	}) {
		t.Errorf("got incorrect SshHostConfig %#v", config.Ec2["foo"])
	}
}

func TestInvalidName(t *testing.T) {
	_, err := loadTestConfig(fullConfig + "\ninvalid-name = bar")
	if err == nil {
		t.Errorf("should have failed")
	}
}
