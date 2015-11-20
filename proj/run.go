package proj

import (
	"flag"
	"log"
	"strings"
)

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

func Main() {
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
