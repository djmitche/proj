package shell

import (
	"io/ioutil"
	"os"
	"path"
	"testing"
)

func TestFindFile(t *testing.T) {
	// make a temporary directory
	tmpdir, err := ioutil.TempDir("", "host_test")
	if err != nil {
		t.Error(err)
		return
	}
	defer func() { os.RemoveAll(tmpdir) }()

	cc := path.Join(tmpdir, "aa", "bb", "cc")
	ff := path.Join(cc, "dd", "ee", "ff")
	err = os.MkdirAll(ff, 0700)
	if err != nil {
		t.Error(err)
		return
	}

	pathname := path.Join(cc, ".proj-testrc")
	err = ioutil.WriteFile(pathname, []byte("test"), 0700)
	if err != nil {
		t.Error(err)
	}

	found, err := findFile(".proj-testrc", ff)
	if err != nil {
		t.Error(err)
	}
	if found != pathname {
		t.Errorf("findFile returned %q", found)
	}

	found, err = findFile("if-this-file-exists-this-test-will-fail", ff)
	if err != nil {
		t.Error(err)
	}
	if found != "" {
		t.Errorf("findFile returned %q", found)
	}
}
