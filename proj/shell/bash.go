package shell

import (
	"fmt"
	"os"
	"os/exec"
	"syscall" // TODO: don't use
)

type bashShell struct {
	rcFilename string
	rcFile     *os.File
}

func newBashShell() (Shell, error) {
	filename := "/tmp/proj-rcfile" // TODO
	file, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY, 0700)
	if err != nil {
		return nil, err
	}
	sh := &bashShell{
		rcFilename: filename,
		rcFile:     file,
	}

	_, err = sh.Write([]byte("[ -f ~/.bashrc ] && . ~/.bashrc\n"))
	if err != nil {
		return nil, err
	}

	return sh, nil
}

func (shell *bashShell) Write(p []byte) (int, error) {
	return shell.rcFile.Write(p)
}

func (shell *bashShell) SetVariable(n, v string) error {
	_, err := shell.rcFile.Write([]byte(fmt.Sprintf("export %s=\"%s\"\n", n, v)))
	if err != nil {
		return err
	}
	return nil
}

func (shell *bashShell) execute() error {
	_, err := shell.Write([]byte(fmt.Sprintf("rm -f \"%s\"\n", shell.rcFilename)))
	if err != nil {
		return err
	}
	err = shell.rcFile.Close()
	if err != nil {
		return err
	}

	shellPath, err := exec.LookPath("bash")
	if err != nil {
		return fmt.Errorf("could not find bash: %s", err)
	}

	return syscall.Exec(shellPath,
		[]string{shellPath, "--rcfile", shell.rcFilename, "-i"}, nil)
}
