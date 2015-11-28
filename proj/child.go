package proj

import (
	"fmt"
	"github.com/djmitche/proj/proj/config"
	"github.com/djmitche/proj/proj/util"
	"log"
	"os"
)

/* child handling */

type Child interface {
	Start(config *config.Config, context Context, path string) error
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

func (child *cdChild) Start(config *config.Config, context Context, path string) error {
	err := os.Chdir(child.dir)
	if err != nil {
		return err
	}

	// re-run from the top, in the same process
	return run(context, child.configFilename, path)
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

// Utility function to re-execute proj in the new environment
// TODO: unused
func localReExecute(context Context, path string) error {
	log.Printf("running %s", os.Args[0])

	// Fork a new child, then write out the context and exit.  This results
	// in a rapid cascade of sub-processes, with ppid=1, but Go doesn't allow
	// raw forks, it seems. (TODO)
	// TODO: just re-run Main?
	r, w, err := os.Pipe()
	if err != nil {
		return err
	}

	args := []string{os.Args[0], "--cfd", "3", path}
	procattr := os.ProcAttr{
		Files: []*os.File{os.Stdin, os.Stdout, os.Stderr, r},
	}
	proc, err := os.StartProcess(args[0], args, &procattr)
	if err != nil {
		return err
	}

	err = r.Close()
	if err != nil {
		return err
	}

	writeContext(context, w)

	// TODO: would rather just exit here, but then the caller forgets about us
	proc.Wait()
	os.Exit(0)

	return nil
}

// Start the child named by `elt`
func StartChild(config *config.Config, context Context, elt string, path string) error {
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

	return child.Start(config, context, path)
}
