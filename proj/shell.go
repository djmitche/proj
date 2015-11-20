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

type bash_shell struct {
	rc_filename string
	rcfile      *os.File
}

func new_bash_shell() Shell {
	filename := "/tmp/proj-rcfile" // TODO
	file, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY, 0700)
	if err != nil {
		log.Panic(err)
	}
	sh := &bash_shell{
		rc_filename: filename,
		rcfile:      file,
	}

	sh.Write([]byte("[ -f ~/.bashrc ] && . ~/.bashrc\n"))

	return sh
}

func (shell *bash_shell) Write(p []byte) (int, error) {
	return shell.rcfile.Write(p)
}

func (shell *bash_shell) SetVariable(n, v string) {
	_, err := shell.rcfile.Write([]byte(fmt.Sprintf("export %s=\"%s\"\n", n, v)))
	if err != nil {
		log.Panic(err)
	}
}

func (shell *bash_shell) execute() {
	shell.Write([]byte(fmt.Sprintf("rm -f \"%s\"\n", shell.rc_filename)))
	shell.rcfile.Close()

	// TODO: search PATH for the shell
	// TODO: execute a shell script
	err := syscall.Exec("/usr/bin/bash", []string{"bash", "--rcfile", shell.rc_filename, "-i"}, nil)
	log.Panic(err)
}

func do_shell(config Config, context Context) {
	log.Printf("do_shell(%+v, %+v)\n", config, context) // TODO

	if context.Shell != "bash" {
		log.Fatalf("unkonwn shell %s", context.Shell)
	}

	shell := new_bash_shell()

	for _, mod := range context.Modifiers {
		mod.Apply(shell)
	}

	shell.execute()
}
