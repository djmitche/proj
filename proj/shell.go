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
	SetVariable(n, v string)

	// Actually execute the shell
	execute()
}

type bashShell struct {
	rcFilename string
	rcFile     *os.File
}

func newBashShell() Shell {
	filename := "/tmp/proj-rcfile" // TODO
	file, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY, 0700)
	if err != nil {
		log.Panic(err)
	}
	sh := &bashShell{
		rcFilename: filename,
		rcFile:     file,
	}

	sh.Write([]byte("[ -f ~/.bashrc ] && . ~/.bashrc\n"))

	return sh
}

func (shell *bashShell) Write(p []byte) (int, error) {
	return shell.rcFile.Write(p)
}

func (shell *bashShell) SetVariable(n, v string) {
	_, err := shell.rcFile.Write([]byte(fmt.Sprintf("export %s=\"%s\"\n", n, v)))
	if err != nil {
		log.Panic(err)
	}
}

func (shell *bashShell) execute() {
	shell.Write([]byte(fmt.Sprintf("rm -f \"%s\"\n", shell.rcFilename)))
	shell.rcFile.Close()

	// TODO: search PATH for the shell
	// TODO: execute a shell script
	err := syscall.Exec("/usr/bin/bash", []string{"bash", "--rcfile", shell.rcFilename, "-i"}, nil)
	log.Panic(err)
}

func doShell(config Config, context Context) {
	log.Printf("doShell(%+v, %+v)\n", config, context) // TODO

	if context.Shell != "bash" {
		log.Fatalf("unkonwn shell %s", context.Shell)
	}

	shell := newBashShell()

	for _, mod := range context.Modifiers {
		mod.Apply(shell)
	}

	shell.execute()
}
