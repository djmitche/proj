package shell

import (
	"fmt"
	"log"
)

type envModifier struct {
	variables map[string]string
}

func newEnvModifier(args interface{}) (Modifier, error) {
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
		//raw:       raw, XXX??
		variables: variables,
	}, nil
}

func (mod *envModifier) Apply(shell Shell) error {
	for n, v := range mod.variables {
		log.Printf("applying %s=%s", n, v)
		// TODO: shell quoting
		err := shell.SetVariable(n, v)
		if err != nil {
			return err
		}
	}
	return nil
}

func init() {
	modifierFactories["env"] = newEnvModifier
}
