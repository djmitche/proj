package proj

import (
	"flag"
	"log"
	"strings"
)

/* main */

func run(context Context, env_config string, path string) {
	log.Printf("run(%#v, %#v, %#v)", context, env_config, path)
	config := load_config(env_config)

	// incorporate the configuration into the accumulated context
	context.Update(config)

	// either start a shell or enter the next path element
	if len(path) == 0 {
		do_shell(config, context)
	} else {
		i := strings.Index(path, "/")
		if i < 0 {
			StartChild(config, context, path, "")
		} else {
			StartChild(config, context, path[:i], path[i+1:])
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

	context := load_context(*cfd)
	run(context, *env_config, path)
}
