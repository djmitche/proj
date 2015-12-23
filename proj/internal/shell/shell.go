package shell

import (
	"fmt"
	"github.com/djmitche/proj/proj/internal/config"
	"log"
	"os"
	"os/exec"
	"syscall" // TODO: don't use
)

// Spawn a new shell in the current directory using the given configuration

func Spawn(hostConfig *config.HostConfig) error {
	log.Printf("Spawn(%+v)\n", hostConfig)

	shellPath, err := exec.LookPath("bash")
	if err != nil {
		return fmt.Errorf("could not find bash: %s", err)
	}

	// TODO: get shellrc from the last child configuration file (pass it along?)
	rcfile := ".projrc"

	var args []string
	args = append(args, shellPath)

	st, err := os.Stat(rcfile)
	if err != nil && st != nil {
		args = append(args, "--rcfile")
		args = append(args, rcfile)
	}

	args = append(args, "-i")

	return syscall.Exec(shellPath, args, os.Environ())
}
