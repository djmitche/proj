package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/kylelemons/go-gypsy/yaml"
	"log"
	"os"
	"path"
	"strings"
	"syscall" // TODO: don't use
)

/* Config handling */

type ConfigElt struct {
	Type string
	Args yaml.Node
}

type Config struct {
	Filename string
	Children map[string]ConfigElt
	Contexts []ConfigElt

	raw yaml.Node
}

func (c Config) String() string {
	return fmt.Sprintf("{\n    Filename: %#v\n    Children: %#v\n    Contexts: %#v\n}",
		c.Filename, c.Children, c.Contexts)
}

func load_config(env_config string) Config {
	var config Config
	var filename string

	if len(env_config) > 0 {
		filename = env_config
	} else {
		wd, err := os.Getwd()
		if err != nil {
			log.Panic(err)
		}
		filename = path.Clean(path.Join(wd, ".proj.yml"))
		if _, err := os.Stat(filename); err != nil {
			log.Printf("fallback, %#v", filename)
			dirname := path.Base(wd)
			filename = path.Clean(path.Join(wd, fmt.Sprintf("../%s-proj.yml", dirname)))
		}
	}

	if _, err := os.Stat(filename); err != nil {
		log.Panic(fmt.Sprintf("Config file '%s' not found", filename))
	}
	config.Filename = filename

	// TODO: load ~/.projrc.yml too
	file, err := yaml.ReadFile(config.Filename)
	if err != nil {
		log.Panic(err)
	}
	config.raw = file.Root

	// parse children
	config.Children = make(map[string]ConfigElt)
	child_node, err := yaml.Child(config.raw, ".children")
	if err == nil {
		for name, value := range child_node.(yaml.Map) {
			val_map := value.(yaml.Map)
			if len(val_map) != 1 {
				log.Panic("malformed configuration 1")
			}
			for typ, args := range val_map {
				config.Children[name] = ConfigElt{
					Type: typ,
					Args: args}
			}
		}
	}

	// parse contexts
	contexts_node, err := yaml.Child(config.raw, ".context")
	if err == nil {
		contexts_list := contexts_node.(yaml.List)
		config.Contexts = make([]ConfigElt, len(contexts_list))
		for i, ctx := range contexts_list {
			ctx_map := ctx.(yaml.Map)
			if len(ctx_map) != 1 {
				log.Panic("malformed configuration 2")
			}
			for typ, args := range ctx_map {
				config.Contexts[i] = ConfigElt{
					Type: typ,
					Args: args}
			}
		}
	} else {
		config.Contexts = make([]ConfigElt, 0)
	}

	return config
}

/* Context handling */

type Context struct {
	Shell     string
	Path      []string
	Modifiers []interface{}
}

func load_context(cfd int) Context {
	if cfd == 0 {
		return Context{
			Shell:     "bash", // TODO from supported_shells
			Path:      []string{},
			Modifiers: make([]interface{}, 0),
		}
	}

	// read from the given file descriptor and close it
	ctxfile := os.NewFile(uintptr(cfd), "ctxfile")
	decoder := json.NewDecoder(ctxfile)
	context := Context{}
	err := decoder.Decode(&context)
	if err != nil {
		log.Panic(err)
	}
	ctxfile.Close()

	log.Printf("got context %#v\n", context)

	return context
}

func write_context(context Context, w *os.File) {
	encoder := json.NewEncoder(w)
	err := encoder.Encode(context)
	if err != nil {
		log.Panic(err)
	}
	w.Close()
}

/* shell handling */

func do_shell(config Config, context Context) {
	log.Printf("do_shell(%+v, %+v)\n", config, context) // TODO

	if context.Shell != "bash" {
		log.Fatalf("unkonwn shell %s", context.Shell)
	}

	// TODO: search PATH for the shell
	// TODO: execute a shell script
	err := syscall.Exec("/usr/bin/bash", []string{"bash"}, nil)
	log.Panic(err)
}

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

/* main */

func run(cfd int, env_config string, path string) {
	config := load_config(env_config)
	context := load_context(cfd)

	// TODO: update context based on config

	// either start a shell or enter the next path element
	if len(path) == 0 {
		do_shell(config, context)
	} else {
		i := strings.Index(path, "/")
		if i < 0 {
			start_child(config, context, path, "")
		} else {
			start_child(config, context, path[:i], path[i+1:])
		}
	}
}

func main() {
	cfd := flag.Int("cfd", 0, "(internal use only)")
	env_config := flag.String("env-config", "", "(internal use only)")

	flag.Parse()

	args := flag.Args()
	if len(args) != 1 {
		log.Panic("Path is required")
	}

	path := args[0]

	run(*cfd, *env_config, path)
}
