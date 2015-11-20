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

func load_config(env_config string) Config {
	var config Config
	var filename string

	if len(env_config) > 0 {
		filename = env_config
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
	root := file.Root

	// parse children
	config.Children = make(map[string]Child)
	child_node, err := yaml.Child(root, ".children")
	if err == nil {
		for name, value := range child_node.(yaml.Map) {
			val_map := value.(yaml.Map)
			if len(val_map) != 1 {
				log.Panic("malformed configuration 1")
			}
			for typ, args := range val_map {
				child := NewChild(typ)
				child.ParseArgs(args)
				config.Children[name] = child
			}
		}
	}

	// parse shell modifiers
	contexts_node, err := yaml.Child(root, ".shell")
	if err == nil {
		yaml_list := contexts_node.(yaml.List)
		contexts := make([]interface{}, yaml_list.Len())
		for i, elt := range yaml_list {
			contexts[i] = yaml_to_json(elt)
		}
		config.Modifiers = contexts
	} else {
		config.Modifiers = make([]interface{}, 0)
	}

	return config
}
