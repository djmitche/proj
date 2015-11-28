package proj

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"strings"
)

/* main */

func run(context Context, configFilename string, path string) error {
	log.Printf("run(%#v, %#v, %#v)", context, configFilename, path)
	config, err := loadConfig(configFilename)
	if err != nil {
		return err
	}

	// incorporate the configuration into the accumulated context
	err = context.Update(config)
	if err != nil {
		return err
	}

	// either start a shell or enter the next path element
	if len(path) == 0 {
		err = doShell(config, context)
	} else {
		i := strings.Index(path, "/")
		if i < 0 {
			err = StartChild(config, context, path, "")
		} else {
			err = StartChild(config, context, path[:i], path[i+1:])
		}
	}

	return err
}

func Main() error {
	verbose := flag.Bool("v", false, "enable verbose logging")
	cfd := flag.Int("cfd", 0, "(internal use only)")
	configFilename := flag.String("config", "", "(internal use only)")

	flag.Parse()

	if !*verbose {
		log.SetOutput(ioutil.Discard)
	}

	args := flag.Args()
	if len(args) != 1 {
		return fmt.Errorf("Path argument is required")
	}

	path := args[0]

	context, err := loadContext(*cfd)
	if err != nil {
		return err
	}
	err = run(context, *configFilename, path)
	if err != nil {
		return err
	}
	return nil
}
