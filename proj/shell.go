package proj

import (
	"fmt"
	"log"
	"os"
	"syscall" // TODO: don't use
)

/* shell handling */

type Shell interface {
	// Set an environment variable in the shell
	SetVariable(n, v string) error

	// Actually execute the shell
	execute() error
}

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

	// TODO: search PATH for the shell
	// TODO: execute a shell script
	return syscall.Exec("/usr/bin/bash", []string{"bash", "--rcfile", shell.rcFilename, "-i"}, nil)
}

func doShell(config Config, context Context) error {
	log.Printf("doShell(%+v, %+v)\n", config, context) // TODO

	if context.Shell != "bash" {
		return fmt.Errorf("unkonwn shell %s", context.Shell)
	}

	shell, err := newBashShell()
	if err != nil {
		return err
	}

	for _, mod := range context.Modifiers {
		err = mod.Apply(shell)
		if err != nil {
			return err
		}
	}

	return shell.execute()
}
