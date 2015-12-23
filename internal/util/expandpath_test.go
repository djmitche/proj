package util

import (
	"os"
	"os/user"
	"testing"
)

func TestExpandPathAbsolute(t *testing.T) {
	if ExpandPath("/foo/bar") != "/foo/bar" {
		t.Fail()
	}
}

func TestExpandPathRelative(t *testing.T) {
	cwd, _ := os.Getwd()
	if ExpandPath("foo/bar") != cwd+"/foo/bar" {
		t.Fail()
	}
}

func TestExpandPathHome(t *testing.T) {
	usr, _ := user.Current()
	if ExpandPath("~/bar") != usr.HomeDir+"/bar" {
		t.Fail()
	}
}
