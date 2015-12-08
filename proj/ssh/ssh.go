package ssh

import (
	"fmt"
	"github.com/djmitche/shquote"
	"log"
	"os"
	"os/exec"
	"syscall"
)

type Config struct {
	// connection information
	User string
	Host string

	// remote proj configuration
	ProjPath       string
	ConfigFilename string
	Path           string
}

// Run proj via SSH
func Run(cfg *Config) error {
	log.Printf("Connecting to %q via SSH", cfg.Host)

	// build an ssh command line
	var sshArgs []string
	sshPath, err := exec.LookPath("ssh")
	if err != nil {
		return fmt.Errorf("'ssh' not found: %s", err)
	}

	sshArgs = append(sshArgs, sshPath)
	sshArgs = append(sshArgs, "-t")
	if cfg.User != "" {
		sshArgs = append(sshArgs, "-l")
		sshArgs = append(sshArgs, cfg.User)
	}
	sshArgs = append(sshArgs, cfg.Host)

	// ssh runs the command by taking all of the arguments ssh itself got,
	// joining them with spaces, and handing them to `sh -c`.  So we need to
	// include quoted strings from here on out.
	sshArgs = append(sshArgs, shquote.Quote(cfg.ProjPath))
	if cfg.ConfigFilename != "" {
		sshArgs = append(sshArgs, "-config")
		sshArgs = append(sshArgs, shquote.Quote(cfg.ConfigFilename))
	}

	// TODO: support running proj in a subdir on the remote system

	// TODO: support setting ForwardAgent

	sshArgs = append(sshArgs, shquote.Quote(cfg.Path))

	// Exec SSH (POSIX only)
	err = syscall.Exec(sshArgs[0], sshArgs, os.Environ())
	return fmt.Errorf("while invoking ssh: %s", err)
}
