package child

import (
	"fmt"
	"os"
	"syscall"
)

func shellChild(info *childInfo) error {
	command := info.childConfig.Shell.Command
	shArgs := []string{"sh", "-c", command}

	err := syscall.Exec("/bin/sh", shArgs, os.Environ())
	return fmt.Errorf("while invoking sh: %s", err)
}

func init() {
	childFuncs["shell"] = shellChild
}
