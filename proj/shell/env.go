package shell

import (
	"fmt"
	"github.com/spf13/cast"
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
		varMap, err := cast.ToStringMapE(args)
		if err != nil {
			return nil, fmt.Errorf("Invalid env shell modifier %j: %s", args, err)
		}
		variables := make(map[string]string)
		for n, v := range varMap {
			var ok bool
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
