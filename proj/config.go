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
	Children  map[string]ChildConfig
	Modifiers []interface{}
}

type ChildConfig struct {
	Type string
	Args interface{}
}

func loadConfig(configFilename string) (*Config, error) {
	var config Config
	var filenames []string

	cwd, err := os.Getwd()
	if err != nil {
		return nil, err
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
		// return a pointer to an empty config
		return &config, nil
	}
	config.Filename = filename

	// TODO: load ~/.projrc.yml too
	file, err := yaml.ReadFile(config.Filename)
	if err != nil {
		return nil, err
	}

	cfgFile, err := yamlToJson(file.Root)
	if err != nil {
		return nil, err
	}

	cfgMap, ok := cfgFile.(map[string]interface{})
	if !ok {
		return nil, err
	}

	// parse children
	config.Children = make(map[string]ChildConfig)
	childrenNode, ok := cfgMap["children"]
	if ok {
		childrenMap, ok := childrenNode.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("`children` must be a map in %q", filename)
		}
		for name, value := range childrenMap {
			childType, args, err := singleKeyMap(value)
			if err != nil {
				return nil, err
			}
			config.Children[name] = ChildConfig{childType, args}
		}
	}

	// parse shell modifiers
	shellNode, ok := cfgMap["shell"]
	if ok {
		config.Modifiers = shellNode.([]interface{})
	} else {
		config.Modifiers = make([]interface{}, 0)
	}

	return &config, nil
}
