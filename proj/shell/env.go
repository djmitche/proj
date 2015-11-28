package shell

import (
	"fmt"
)

type envModifier struct {
	variables map[string]string
}

func (mod *envModifier) Modify(shell Shell) error {
	for n, v := range mod.variables {
		err := shell.SetVariable(n, v)
		if err != nil {
			return err
		}
	}
	return nil
}

func init() {
	modifierFactories["env"] = func(args interface{}) (Modifier, error) {
		varMap, ok := args.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("Invalid env shell modifier %j", args)
		}
		variables := make(map[string]string)
		for n, v := range varMap {
			variables[n], ok = v.(string)
			if !ok {
				return nil, fmt.Errorf("Invalid env shell modifier %j", args)
			}
		}

		return &envModifier{
			variables: variables,
		}, nil
	}

}
