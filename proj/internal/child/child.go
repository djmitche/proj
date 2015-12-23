// The proj/child package manages traversing into child projects.
package child

import (
	"fmt"
	"github.com/djmitche/proj/proj/internal/config"
	"log"
	"os"
	"path"
)

// a function to recursively "run" proj in a new project
type recurseFunc func(path string) error

// information that is useful to each child function
type childInfo struct {
	// child configuration
	childConfig *config.ChildConfig

	// host configuration
	hostConfig *config.HostConfig

	// remaining proj path after this child
	path string

	// a pointer to the top-level "run" function, useful for recursing into a child
	// in the same process
	recurse recurseFunc
}

// a child function, implemented in other files in this package
type childFunc func(info *childInfo) error

// the map of all known child functions
var childFuncs map[string]childFunc = make(map[string]childFunc)

// load a child configuration given the child project name
func loadChildConfigFor(elt string) (*config.ChildConfig, error) {
	// try .proj/<child>.cfg
	configFilename := path.Join(".proj", elt+".cfg")
	st, err := os.Stat(configFilename)
	if err == nil {
		return config.LoadChildConfig(configFilename)
	}

	// try a simple subdirectory
	st, err = os.Stat(elt)
	if err == nil && st != nil && st.IsDir() {
		cfg := config.ChildConfig{Type: "cd"}
		cfg.Cd.Dir = elt
		return &cfg, nil
	}

	return nil, fmt.Errorf("No such child %s", elt)
}

// Start the child named by `elt`
func StartChild(hostConfig *config.HostConfig, elt string, path string, recurse recurseFunc) error {
	log.Printf("startChild(%+v, %+v, %+v, %+v)\n", hostConfig, elt, path)

	childConfig, err := loadChildConfigFor(elt)
	if err != nil {
		return fmt.Errorf("No such child %s", elt)
	}

	// TODO: apply common stuff here: prepend

	f, ok := childFuncs[childConfig.Type]
	if !ok {
		return fmt.Errorf("No such child type %s", childConfig.Type)
	}

	err = f(&childInfo{
		childConfig: childConfig,
		hostConfig:  hostConfig,
		path:        path,
		recurse:     recurse,
	})
	if err != nil {
		return fmt.Errorf("while starting child %q: %s", elt, err)
	}
	return nil
}
