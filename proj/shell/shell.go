package shell

import (
	"fmt"
	"log"
)

/* shell handling */

type Shell interface {
	// Get the shell type ("bash", "tcsh", etc.).  Modifiers may use this to
	// determine what input they should provide.  It's not terribly important
	// whether a modifier is aware of the different shell types or not.
	Type() string

	// Set an environment variable in the shell
	SetVariable(n, v string) error

	// Source a shell script (which must be appropriate for this shell type)
	Source(file string) error

	// Actually execute the shell
	execute() error
}

// Spawn a new shell using the given Context
func Spawn(context *Context) error {
	log.Printf("Spawn(%+v)\n", context)

	if context.Shell != "bash" {
		return fmt.Errorf("unkonwn shell %s", context.Shell)
	}

	shell, err := newBashShell()
	if err != nil {
		return fmt.Errorf("while creating new shell: %s", err)
	}

	for _, mod := range context.Modifiers {
		err = mod.Modify(shell)
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
