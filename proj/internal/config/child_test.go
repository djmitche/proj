package config

import (
	"io/ioutil"
	"os"
	"testing"
)

func loadTestChildConfig(config string) (*ChildConfig, error) {
	// make a temporary file
	tmpfile, err := ioutil.TempFile("", "host_test")
	if err != nil {
		return nil, err
	}
	filename := tmpfile.Name()
	defer func() { os.RemoveAll(filename) }()

	// write out the contents
	err = ioutil.WriteFile(filename, []byte(config), 0700)
	if err != nil {
		return nil, err
	}

	// load the config
	return LoadChildConfig(filename)
}

func TestSimpleConfig(t *testing.T) {
	config, err := loadTestChildConfig("[ssh]\nhost = foo\n")
	if err != nil {
		t.Log(err)
		t.FailNow()
	}

	if config.Type != "ssh" || config.Ssh.Host != "foo" {
		t.Errorf("That came out wrong: %#v", config)
	}
}
