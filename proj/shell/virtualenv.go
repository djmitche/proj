package shell

import (
	"fmt"
	"path/filepath"
)

type virtualenvModifier struct {
	dir string
}

func (mod *virtualenvModifier) Modify(shell Shell) error {
	switch shell.Type() {
	case "bash":
		activate, err := filepath.Abs(filepath.Join(mod.dir, "bin", "activate"))
		if err != nil {
			return err
		}
		return shell.Source(activate)
	default:
		return fmt.Errorf("Don't know how to activate virtualenv for %s", shell.Type())
	}
}

func init() {
	modifierFactories["virtualenv"] = func(args interface{}) (Modifier, error) {
		dir, ok := args.(string)
		if !ok {
			return nil, fmt.Errorf("Invalid virtualenv directory %j", args)
		}
		return &virtualenvModifier{
			dir: dir,
		}, nil
	}

}
