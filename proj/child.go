package proj

import (
	"log"
	"os"
)

/* child handling */

type Child interface {
	ParseArgs(args interface{})
	Start(config Config, context Context, path string)
}

type childFactory func() Child

var childFactories map[string]childFactory = make(map[string]childFactory)

type cdChild struct {
	dir        string
	env_config string
}

func (child *cdChild) ParseArgs(args interface{}) {
	node, ok := default_child(args, "dir")
	if !ok {
		log.Panic("no dir specified")
	}
	child.dir, ok = node.(string)
	if !ok {
		log.Panic("child dir is not a string")
	}

	args_map, ok := args.(map[string]interface{})
	if ok {
		config_arg, ok := args_map["config"]
		if ok {
			config_arg_str, ok := config_arg.(string)
			if ok {
				child.env_config = config_arg_str
			} else {
				log.Panic("config should be a string")
			}
		}
	}
}

func (child *cdChild) Start(config Config, context Context, path string) {
	err := os.Chdir(child.dir)
	if err != nil {
		log.Panic(err)
	}

	// re-run from the top, in the same process
	run(context, child.env_config, path)
}

func init() {
	childFactories["cd"] = func() Child { return &cdChild{} }
}

func NewChild(child_type string) Child {
	factory, ok := childFactories[child_type]
	if !ok {
		log.Fatalf("No such child type %s", child_type)
	}
	return factory()
}

// Utility function to re-execute proj in the new environment
// TODO: unused
func local_reexec(context Context, path string) {
	log.Printf("running %s", os.Args[0])

	// Fork a new child, then write out the context and exit.  This results
	// in a rapid cascade of sub-processes, with ppid=1, but Go doesn't allow
	// raw forks, it seems. (TODO)
	// TODO: just re-run Main?
	r, w, err := os.Pipe()
	if err != nil {
		log.Panic(err)
	}

	args := []string{os.Args[0], "--cfd", "3", path}
	procattr := os.ProcAttr{
		Files: []*os.File{os.Stdin, os.Stdout, os.Stderr, r},
	}
	proc, err := os.StartProcess(args[0], args, &procattr)
	if err != nil {
		log.Panic(err)
	}

	err = r.Close()
	if err != nil {
		log.Panic(err)
	}

	write_context(context, w)

	// TODO: would rather just exit here, but then the caller forgets about us
	proc.Wait()
	os.Exit(0)
}

// Start the child named by `elt`
func StartChild(config Config, context Context, elt string, path string) {
	log.Printf("start_child(%+v, %+v, %+v, %+v)\n", config, context, elt, path)
	child, ok := config.Children[elt]
	if !ok {
		log.Fatalf("No such child %s", elt)
	}

	// add the element to the context to be handed to the child
	context.Path = append(context.Path, elt)

	child.Start(config, context, path)
}
