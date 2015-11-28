package shell

import (
	"fmt"
	"io/ioutil"
	"log"
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

	_, err = sh.Write([]byte("[ -f ~/.bashrc ] && . ~/.bashrc\nset -e\n"))
	if err != nil {
		return nil, err
	}

	return sh, nil
}

func (shell *bashShell) Type() string {
	return "bash"
}

func (shell *bashShell) Write(p []byte) (int, error) {
	return shell.rcFile.Write(p)
}

func (shell *bashShell) SetVariable(n, v string) error {
	// TODO: shell quoting
	_, err := shell.rcFile.Write([]byte(fmt.Sprintf("export %s=\"%s\"\n", n, v)))
	if err != nil {
		return err
	}
	return nil
}

func (shell *bashShell) Source(file string) error {
	// TODO: shell quoting
	_, err := shell.rcFile.Write([]byte(fmt.Sprintf("source \"%s\"\n", file)))
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

	// log the rcfile contents (ignoring errors)
	rcFileReader, err := os.OpenFile(shell.rcFilename, os.O_RDONLY, 0)
	if err != nil {
		panic(err)
	}
	content, _ := ioutil.ReadAll(rcFileReader)
	rcFileReader.Close()
	log.Printf("rcfile contents:\n====\n%s====", content)

	shellPath, err := exec.LookPath("bash")
	if err != nil {
		return fmt.Errorf("could not find bash: %s", err)
	}

	return syscall.Exec(shellPath,
		[]string{shellPath, "--rcfile", shell.rcFilename, "-i"}, os.Environ())
}
