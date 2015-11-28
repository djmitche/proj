package child

import (
	"fmt"
	"github.com/djmitche/proj/proj/config"
	"github.com/djmitche/proj/proj/shell"
	"log"
)

type recurseFunc func(context shell.Context, configFilename string, path string) error

// information that is useful to each child function
type childInfo struct {
	// the name of the child type
	childType string

	// the args in JSON format
	args interface{}

	// current configuration
	config *config.Config

	// current shell contexst
	context shell.Context

	// remaining proj path after this child
	path string

	// a pointer to the top-level "run" function, useful for recursing into a child
	// in the same process
	recurse recurseFunc
}

type childFunc func(info *childInfo) error

var childFuncs map[string]childFunc = make(map[string]childFunc)

// Start the child named by `elt`
func StartChild(config *config.Config, context shell.Context, elt string, path string, recurse recurseFunc) error {
	log.Printf("startChild(%+v, %+v, %+v, %+v)\n", config, context, elt, path)
	childConfig, ok := config.Children[elt]
	if !ok {
		return fmt.Errorf("No such child %s", elt)
	}

	f, ok := childFuncs[childConfig.Type]
	if !ok {
		return fmt.Errorf("No such child type %s", childConfig.Type)
	}

	err := f(&childInfo{
		childType: childConfig.Type,
		args:      childConfig.Args,
		config:    config,
		context:   context,
		path:      path,
		recurse:   recurse,
	})
	if err != nil {
		return fmt.Errorf("while starting child %q: %s", elt, err)
	}
	return nil
}
