package proj

import (
	"log"
	"syscall" // TODO: don't use
)

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
