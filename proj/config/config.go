package config

import (
	"fmt"
	"github.com/djmitche/proj/proj/util"
	"github.com/spf13/viper"
	"log"
	"os"
	"path"
)

/* Config handling */

type Config struct {
	File      *viper.Viper
	Children  map[string]ChildConfig
	Modifiers []interface{}
}

type ChildConfig struct {
	Type string
	Args interface{}
}

// Load the proj configuration for the current directory.  This will come from
// an explicitly specified configuration file, or from `.proj.yml`, or
// `../<dirname>-proj.yml`.  If no configuration is found, LoadProjConfig will
// return an empty configuration (common for "leaf" projects).
func LoadProjConfig(configFilename string) (*Config, error) {
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
		return &Config{}, nil
	}

	return loadProjectConfigFromFile(filename)
}

func loadProjectConfigFromFile(filename string) (*Config, error) {
	var config Config

	// load the config file with Viper
	v := viper.New()
	v.SetConfigFile(filename)
	err := v.ReadInConfig()
	if err != nil {
		return nil, fmt.Errorf("parsing %q: %s", filename, err)
	}
	config.File = v

	// parse children
	config.Children = make(map[string]ChildConfig)
	if v.IsSet("children") {
		for name, value := range v.GetStringMap("children") {
			childType, args, err := util.SingleKeyMap(value)
			if err != nil {
				return nil, fmt.Errorf("child %q in %q: %s", name, filename, err)
			}
			config.Children[name] = ChildConfig{childType, args}
		}
	}

	// parse shell modifiers
	if v.IsSet("shell") {
		modifiers := v.Get("shell")
		var ok bool
		config.Modifiers, ok = modifiers.([]interface{})
		if !ok {
			return nil, fmt.Errorf("'shell' is not a list")
		}
	} else {
		config.Modifiers = make([]interface{}, 0)
	}

	return &config, nil
}
