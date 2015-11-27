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

func loadConfig(envConfig string) Config {
	var config Config
	var filename string

	if len(envConfig) > 0 {
		filename = envConfig
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

	cfgFile := yamlToJson(file.Root)
	cfgMap, ok := cfgFile.(map[string]interface{})
	if !ok {
		log.Panic(err)
	}

	// parse children
	config.Children = make(map[string]Child)
	childrenNode, ok := cfgMap["children"]
	if ok {
		childrenMap, ok := childrenNode.(map[string]interface{})
		if !ok {
			log.Fatal("`children` must be a map")
		}
		for name, value := range childrenMap {
			childType, args, err := singleKeyMap(value)
			if err != nil {
				log.Panic(err)
			}
			child := NewChild(childType)
			child.ParseArgs(args)
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

	return config
}
