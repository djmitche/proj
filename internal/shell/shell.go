package shell

import (
	"fmt"
	"github.com/djmitche/proj/internal/config"
	"log"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"syscall" // TODO: don't use
)

// Find a file with the given filename in the given directory or some parent
// directory, or empty string if not found
func findFile(filename, dir string) (string, error) {
	dir, err := filepath.Abs(dir)
	if err != nil {
		return "", err
	}

	for dir != "/" {
		pathname := path.Join(dir, filename)
		st, err := os.Stat(pathname)
		if err == nil && st != nil {
			return pathname, nil
		}
		dir = path.Dir(dir)
	}

	return "", nil
}

// Spawn a new shell in the current directory using the given configuration
func Spawn(hostConfig *config.HostConfig) error {
	log.Printf("Spawn(%+v)\n", hostConfig)

	shellPath, err := exec.LookPath("bash")
	if err != nil {
		return fmt.Errorf("could not find bash: %s", err)
	}

	// TODO: get shellrc from the last child configuration file (pass it along?)
	dir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("could not get current directory: %s", err)
	}

	rcfile := hostConfig.Shell.Rcfile
	if rcfile == "" {
		rcfile = ".projrc"
	}
	rcfile, err = findFile(rcfile, dir)
	if err != nil {
		return fmt.Errorf("while searching for rcfile: %s", err)
	}

	var args []string
	args = append(args, shellPath)

	if rcfile != "" {
		args = append(args, "--rcfile")
		args = append(args, rcfile)
	}

	args = append(args, "-i")

	log.Println(args)

	return syscall.Exec(shellPath, args, os.Environ())
}
