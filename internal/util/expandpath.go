package util

import (
	"os"
	"os/user"
	"path"
)

// Expand a relative path to an absolute path, substituting an initial "~/" for
// the user's home directory.  Note that "~username" is not supported.
func ExpandPath(pth string) string {
	if pth[:2] == "~/" {
		usr, err := user.Current()
		if err == nil {
			pth = path.Join(usr.HomeDir, pth[2:])
		}
	}

	if !path.IsAbs(pth) {
		cwd, err := os.Getwd()
		if err == nil {
			pth = path.Join(cwd, pth)
		}
	}

	return pth
}
