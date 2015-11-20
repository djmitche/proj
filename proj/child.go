package proj

import (
	"github.com/kylelemons/go-gypsy/yaml"
	"log"
	"os"
)

/* child handling */

func local_reexec(context Context, path string) {
	log.Printf("running %s", os.Args[0])

	// Fork a new child, then write out the context and exit.  This results
	// in a rapid cascade of sub-processes, with ppid=1, but Go doesn't allow
	// raw forks, it seems. (TODO)
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

// utility function to get a child of a YAML node, or if the YAML node is not a
// Map, assume that it is the expected value.  This allows things like "cd:
// somedir" as a shorthand for "cd: dir: somedir".
func default_child(args yaml.Node, key string) (yaml.Node, error) {
	_, ok := args.(yaml.Map)
	if !ok {
		return args, nil
	} else {
		return yaml.Child(args, key)
	}
}

func start_child_cd(config Config, context Context, child ConfigElt, path string) {
	// TODO: handle Args being a string
	dir_node, err := default_child(child.Args, "dir")
	if err != nil {
		log.Panic(err)
	}
	// TODO: handle not a scalar
	dir_scalar, ok := dir_node.(yaml.Scalar)
	if !ok {
		log.Panic("invalid directory %#v", dir_node)
	}
	dir := dir_scalar.String()

	// TODO: handle 'config' option, sending --env-config

	err = os.Chdir(dir)
	if err != nil {
		log.Panic(err)
	}

	local_reexec(context, path)
}

func start_child(config Config, context Context, elt string, path string) {
	log.Printf("start_child(%+v, %+v, %+v, %+v)\n", config, context, elt, path)
	child, ok := config.Children[elt]
	if !ok {
		log.Fatalf("No such child %s", elt)
	}

	// add the element to the context to be handed to the child
	context.Path = append(context.Path, elt)

	if child.Type == "cd" {
		start_child_cd(config, context, child, path)
	} else {
		log.Panic("unknown child or child type")
	}
}
