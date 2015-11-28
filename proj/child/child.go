// The proj/child package manages traversing the proj path into child projects.
// For example, the proj path `work/openstack/ironic` requires first traversing
// to the `work` project, then `openstack`, and then `ironic`, collecting context
// along the way.
//
// Children are configured for each project like this:
//
// children:
//     openstack:
//         cd: openstack/src
//
// This says that the child named `openstack` has type `cd` with argument
// `openstack/src`.  Child type implementations get handed a wealth of information
// and can use it to re-run `proj` in the child project, in whatever way is most
// appropriate.
package child

import (
	"fmt"
	"github.com/djmitche/proj/proj/config"
	"github.com/djmitche/proj/proj/shell"
	"log"
)

type recurseFunc func(context *shell.Context, configFilename string, path string) error

// information that is useful to each child function
type childInfo struct {
	// the name of the child type
	childType string

	// the args in JSON format
	args interface{}

	// current configuration
	config *config.Config

	// current shell contexst
	context *shell.Context

	// remaining proj path after this child
	path string

	// a pointer to the top-level "run" function, useful for recursing into a child
	// in the same process
	recurse recurseFunc
}

type childFunc func(info *childInfo) error

var childFuncs map[string]childFunc = make(map[string]childFunc)

// Start the child named by `elt` in the current project's configuration, based on the given context.
func StartChild(config *config.Config, context *shell.Context, elt string, path string, recurse recurseFunc) error {
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
