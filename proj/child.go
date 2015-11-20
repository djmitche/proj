package proj

import (
	"github.com/kylelemons/go-gypsy/yaml"
	"log"
	"os"
)

/* child handling */

type Child interface {
	ParseArgs(args yaml.Node)
	Start(config Config, context Context, path string)
}

type cdChild struct {
	dir        string
	env_config string
}

func (child *cdChild) ParseArgs(args yaml.Node) {
	node, err := default_child(args, "dir")
	if err != nil {
		log.Panic(err)
	}
	child.dir = node_string(node)

	node, err = yaml.Child(args, "config")
	if err == nil {
		child.env_config = node_string(node)
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

func NewChild(child_type string) Child {
	// TODO: use a map and functors
	if child_type == "cd" {
		return &cdChild{}
	} else {
		log.Fatalf("No such child type %s", child_type)
	}
	return nil
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
