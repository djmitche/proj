package proj

import (
	"flag"
	"fmt"
	"github.com/djmitche/proj/proj/internal/child"
	"github.com/djmitche/proj/proj/internal/config"
	"github.com/djmitche/proj/proj/internal/shell"
	"io/ioutil"
	"log"
	"strings"
)

/* main */

func run(path string) error {
	log.Printf("run(%#v)", path)

	hostConfig, err := config.LoadHostConfig()
	if err != nil {
		return err
	}

	// either start a shell or enter the next path element
	if len(path) == 0 {
		err = shell.Spawn(nil)
	} else {
		i := strings.Index(path, "/")
		if i < 0 {
			err = child.StartChild(hostConfig, path, "", run)
		} else {
			err = child.StartChild(hostConfig, path[:i], path[i+1:], run)
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
