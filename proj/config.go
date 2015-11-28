package proj

import (
	"fmt"
	"github.com/kylelemons/go-gypsy/yaml"
	"log"
	"os"
	"path"
)

/* Config handling */

type Config struct {
	Filename  string
	Children  map[string]Child
	Modifiers []interface{}
}

func (c Config) String() string {
	return fmt.Sprintf("{\n    Filename: %#v\n    Children: %#v\n    Modifiers: %#v\n}",
		c.Filename, c.Children, c.Modifiers)
}

func loadConfig(configFilename string) (Config, error) {
	var config Config
	var filenames []string

	cwd, err := os.Getwd()
	if err != nil {
		return Config{}, err
	}

	if configFilename != "" {
		filenames = []string{configFilename}
	} else {
		filenames = append(filenames,
			path.Clean(path.Join(cwd, ".proj.yml")))

		dirname := path.Base(cwd)
		filenames = append(filenames,
			path.Clean(path.Join(cwd, fmt.Sprintf("../%s-proj.yml", dirname))))
	}

	var filename string
	for _, fn := range filenames {
		if _, err := os.Stat(fn); err == nil {
			filename = fn
			break
		}
	}
	if filename == "" {
		log.Printf("WARNING: no config file found for %q", cwd)
		return Config{}, nil
	}
	config.Filename = filename

	// TODO: load ~/.projrc.yml too
	file, err := yaml.ReadFile(config.Filename)
	if err != nil {
		return Config{}, err
	}

	cfgFile, err := yamlToJson(file.Root)
	if err != nil {
		return Config{}, err
	}

	cfgMap, ok := cfgFile.(map[string]interface{})
	if !ok {
		return Config{}, err
	}

	// parse children
	config.Children = make(map[string]Child)
	childrenNode, ok := cfgMap["children"]
	if ok {
		childrenMap, ok := childrenNode.(map[string]interface{})
		if !ok {
			return Config{}, fmt.Errorf("`children` must be a map in %q", filename)
		}
		for name, value := range childrenMap {
			childType, args, err := singleKeyMap(value)
			if err != nil {
				return Config{}, err
			}
			child, err := NewChild(childType)
			if err != nil {
				return Config{}, fmt.Errorf("parsing child %q in %q: %s", name, filename, err)
			}
			err = child.ParseArgs(args)
			if err != nil {
				return Config{}, err
			}
			config.Children[name] = child
		}
	}

	// parse shell modifiers
	shellNode, ok := cfgMap["shell"]
	if ok {
		config.Modifiers = shellNode.([]interface{})
	} else {
		config.Modifiers = make([]interface{}, 0)
	}

	return config, nil
}
