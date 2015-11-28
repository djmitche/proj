package shell

import (
	"fmt"
	"github.com/djmitche/proj/proj/config"
	"log"
)

/* shell handling */

type Shell interface {
	// Set an environment variable in the shell
	SetVariable(n, v string) error

	// Actually execute the shell
	execute() error
}

func Spawn(config *config.Config, context *Context) error {
	log.Printf("Spawn(%+v, %+v)\n", config, context)

	if context.Shell != "bash" {
		return fmt.Errorf("unkonwn shell %s", context.Shell)
	}

	shell, err := newBashShell()
	if err != nil {
		return fmt.Errorf("while creating new shell: %s", err)
	}

	for _, mod := range context.Modifiers {
		err = mod.Apply(shell)
		if err != nil {
			return fmt.Errorf("while applying modifier %q to shell: %s",
				mod, err)
		}
	}

	err = shell.execute()
	if err != nil {
		return fmt.Errorf("while executing shell: %s", err)
	}
	return nil
}
