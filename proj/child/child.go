package child

import (
	"fmt"
	"github.com/djmitche/proj/proj/config"
	"github.com/djmitche/proj/proj/shell"
	"github.com/djmitche/proj/proj/util"
	"log"
	"os"
)

type recurseFunc func(context shell.Context, configFilename string, path string) error

/* child handling */

type Child interface {
	Start(config *config.Config, context shell.Context, path string, recurse recurseFunc) error
}

type childFactory func(interface{}) (Child, error)

var childFactories map[string]childFactory = make(map[string]childFactory)

type cdChild struct {
	dir            string
	configFilename string
}

func newCdChild(args interface{}) (Child, error) {
	var child cdChild

	node, ok := util.DefaultChild(args, "dir")
	if !ok {
		return nil, fmt.Errorf("no dir specified")
	}
	child.dir, ok = node.(string)
	if !ok {
		return nil, fmt.Errorf("child dir is not a string")
	}

	argsMap, ok := args.(map[string]interface{})
	if ok {
		configArg, ok := argsMap["config"]
		if ok {
			configArgStr, ok := configArg.(string)
			if ok {
				child.configFilename = configArgStr
			} else {
				return nil, fmt.Errorf("config should be a string")
			}
		}
	}

	return &child, nil
}

func (child *cdChild) Start(config *config.Config, context shell.Context, path string, recurse recurseFunc) error {
	err := os.Chdir(child.dir)
	if err != nil {
		return err
	}

	// re-run from the top, in the same process
	return recurse(context, child.configFilename, path)
}

func init() {
	childFactories["cd"] = newCdChild
}

func newChild(childType string, args interface{}) (Child, error) {
	factory, ok := childFactories[childType]
	if !ok {
		return nil, fmt.Errorf("No such child type %s", childType)
	}
	return factory(args)
}

// Start the child named by `elt`
func StartChild(config *config.Config, context shell.Context, elt string, path string, recurse recurseFunc) error {
	log.Printf("startChild(%+v, %+v, %+v, %+v)\n", config, context, elt, path)
	childConfig, ok := config.Children[elt]
	if !ok {
		return fmt.Errorf("No such child %s", elt)
	}

	child, err := newChild(childConfig.Type, childConfig.Args)
	if err != nil {
		return err
	}

	// add the path element to the context to be handed to the child
	context.Path = append(context.Path, elt)

	return child.Start(config, context, path, recurse)
}
