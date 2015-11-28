package proj

import (
	"fmt"
	"log"
	"os"
)

/* child handling */

type Child interface {
	ParseArgs(args interface{}) error
	Start(config Config, context Context, path string) error
}

type childFactory func() Child

var childFactories map[string]childFactory = make(map[string]childFactory)

type cdChild struct {
	dir       string
	envConfig string
}

func (child *cdChild) ParseArgs(args interface{}) error {
	node, ok := defaultChild(args, "dir")
	if !ok {
		return fmt.Errorf("no dir specified")
	}
	child.dir, ok = node.(string)
	if !ok {
		return fmt.Errorf("child dir is not a string")
	}

	argsMap, ok := args.(map[string]interface{})
	if ok {
		configArg, ok := argsMap["config"]
		if ok {
			configArgStr, ok := configArg.(string)
			if ok {
				child.envConfig = configArgStr
			} else {
				return fmt.Errorf("config should be a string")
			}
		}
	}

	return nil
}

func (child *cdChild) Start(config Config, context Context, path string) error {
	err := os.Chdir(child.dir)
	if err != nil {
		return err
	}

	// re-run from the top, in the same process
	return run(context, child.envConfig, path)
}

func init() {
	childFactories["cd"] = func() Child { return &cdChild{} }
}

func NewChild(childType string) (Child, error) {
	factory, ok := childFactories[childType]
	if !ok {
		return nil, fmt.Errorf("No such child type %s", childType)
	}
	return factory(), nil
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
func StartChild(config Config, context Context, elt string, path string) error {
	log.Printf("startChild(%+v, %+v, %+v, %+v)\n", config, context, elt, path)
	child, ok := config.Children[elt]
	if !ok {
		return fmt.Errorf("No such child %s", elt)
	}

	// add the path element to the context to be handed to the child
	context.Path = append(context.Path, elt)

	return child.Start(config, context, path)
}
