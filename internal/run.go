package proj

import (
	"flag"
	"fmt"
	"github.com/djmitche/proj/internal/child"
	"github.com/djmitche/proj/internal/config"
	"github.com/djmitche/proj/internal/shell"
	"io/ioutil"
	"log"
	"strings"
)

/* main */

func run(path string) error {
	var err error

	log.Printf("run(%q)", path)

	hostConfig, err := config.LoadHostConfig()
	if err != nil {
		return err
	}

	// either start a shell or enter the next path element
	if len(path) > 0 {
		var elt, remaining string
		i := strings.Index(path, "/")
		if i < 0 {
			elt, remaining = path, ""
		} else {
			elt, remaining = path[:i], path[i+1:]
		}
		err = child.StartChild(hostConfig, elt, remaining, run)
		if err != nil {
			fmt.Printf("While starting child %s: %s\n", elt, err)
			// try to start a shell here
			err = shell.Spawn(hostConfig)
		}
	} else {
		// if there is child config for `DEFAULT`, start it
		if child.Exists(hostConfig, "DEFAULT") {
			err = child.StartChild(hostConfig, "DEFAULT", "", run)
			// falling back to a local shell
			if err != nil {
				err = shell.Spawn(hostConfig)
			}
		} else {
			err = shell.Spawn(hostConfig)
		}
	}

	return err
}

func Main() error {
	verbose := flag.Bool("v", false, "enable verbose logging")

	flag.Parse()

	if !*verbose {
		log.SetOutput(ioutil.Discard)
	}

	args := flag.Args()
	if len(args) != 1 {
		return fmt.Errorf("Path argument is required")
	}

	path := args[0]

	err := run(path)
	if err != nil {
		return err
	}

	return nil
}
