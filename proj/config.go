package proj

import (
	"fmt"
	"github.com/kylelemons/go-gypsy/yaml"
	"log"
	"os"
	"path"
)

/* Config handling */

type ConfigElt struct {
	Type string
	Args yaml.Node
}

type Config struct {
	Filename string
	Children map[string]ConfigElt
	Contexts []ConfigElt

	raw yaml.Node
}

func (c Config) String() string {
	return fmt.Sprintf("{\n    Filename: %#v\n    Children: %#v\n    Contexts: %#v\n}",
		c.Filename, c.Children, c.Contexts)
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
	config.raw = file.Root

	// parse children
	config.Children = make(map[string]ConfigElt)
	child_node, err := yaml.Child(config.raw, ".children")
	if err == nil {
		for name, value := range child_node.(yaml.Map) {
			val_map := value.(yaml.Map)
			if len(val_map) != 1 {
				log.Panic("malformed configuration 1")
			}
			for typ, args := range val_map {
				config.Children[name] = ConfigElt{
					Type: typ,
					Args: args}
			}
		}
	}

	// parse contexts
	contexts_node, err := yaml.Child(config.raw, ".context")
	if err == nil {
		contexts_list := contexts_node.(yaml.List)
		config.Contexts = make([]ConfigElt, len(contexts_list))
		for i, ctx := range contexts_list {
			ctx_map := ctx.(yaml.Map)
			if len(ctx_map) != 1 {
				log.Panic("malformed configuration 2")
			}
			for typ, args := range ctx_map {
				config.Contexts[i] = ConfigElt{
					Type: typ,
					Args: args}
			}
		}
	} else {
		config.Contexts = make([]ConfigElt, 0)
	}

	return config
}
