package ssh

import (
	"fmt"
	"github.com/djmitche/proj/internal/config"
	"github.com/djmitche/shquote"
	"log"
	"os"
	"os/exec"
	"syscall"
)

// Run proj via SSH
func Run(hostname string, sshConfig *config.SshCommonConfig, path string) error {
	log.Printf("Connecting to %q via SSH", hostname)

	// build an ssh command line
	var sshArgs []string
	sshPath, err := exec.LookPath("ssh")
	if err != nil {
		return fmt.Errorf("'ssh' not found: %s", err)
	}

	sshArgs = append(sshArgs, sshPath)
	sshArgs = append(sshArgs, "-t")

	if sshConfig.User != "" {
		sshArgs = append(sshArgs, "-l")
		sshArgs = append(sshArgs, sshConfig.User)
	}

	if sshConfig.Forward_Agent {
		sshArgs = append(sshArgs, "-o")
		sshArgs = append(sshArgs, "ForwardAgent=yes")
	}

	if sshConfig.Ignore_Known_Hosts {
		sshArgs = append(sshArgs, "-o")
		sshArgs = append(sshArgs, "StrictHostKeyChecking=no")
		sshArgs = append(sshArgs, "-o")
		sshArgs = append(sshArgs, "UserKnownHostsFile=/dev/null")
	}

	sshArgs = append(sshArgs, hostname)

	projPath := sshConfig.Proj_Path
	if projPath == "" {
		projPath = "proj"
	}

	// ssh runs the command by taking all of the arguments ssh itself got,
	// joining them with spaces, and handing them to `sh -c`.  So we need to
	// include quoted strings from here on out.
	sshArgs = append(sshArgs, shquote.Quote(projPath))

	// TODO: support running proj in a subdir on the remote system

	sshArgs = append(sshArgs, shquote.Quote(path))

	log.Printf("SSH command line: %q", sshArgs)

	// Exec SSH (POSIX only)
	err = syscall.Exec(sshArgs[0], sshArgs, os.Environ())
	return fmt.Errorf("while invoking ssh: %s", err)
}
