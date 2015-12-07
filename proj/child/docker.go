package child

import (
	"fmt"
	"github.com/djmitche/proj/proj/util"
	"log"
	"os"
	"os/exec"
	"syscall"
)

type dockerConfig struct {
	image          string
	user           string
	dir            string
	projPath       string
	configFilename string
	args           []string
}

func run(cfg *dockerConfig, info *childInfo) error {
	log.Printf("Starting docker container from %s", cfg.image)

	// build an docker command line
	var dockerArgs []string
	dockerPath, err := exec.LookPath("docker")
	if err != nil {
		return fmt.Errorf("'docker' not found: %s", err)
	}

	dockerArgs = append(dockerArgs, dockerPath)
	dockerArgs = append(dockerArgs, "run")
	dockerArgs = append(dockerArgs, "--rm") // remove container on exit
	dockerArgs = append(dockerArgs, "-ti")  // terminal, interactive
	if cfg.user != "" {
		dockerArgs = append(dockerArgs, "-u")
		dockerArgs = append(dockerArgs, cfg.user)
	}
	if cfg.dir != "" {
		dockerArgs = append(dockerArgs, "-w")
		dockerArgs = append(dockerArgs, cfg.dir)
	}

	if cfg.args != nil {
		for _, arg := range cfg.args {
			dockerArgs = append(dockerArgs, arg)
		}
	}

	dockerArgs = append(dockerArgs, cfg.image)
	if cfg.projPath != "" {
		dockerArgs = append(dockerArgs, cfg.projPath)
	} else {
		dockerArgs = append(dockerArgs, "proj")
	}
	if cfg.configFilename != "" {
		dockerArgs = append(dockerArgs, "-config")
		dockerArgs = append(dockerArgs, cfg.configFilename)
	}

	// TODO: support running proj in a subdir in the image

	if info.path == "" {
		dockerArgs = append(dockerArgs, "''")
	} else {
		dockerArgs = append(dockerArgs, info.path)
	}

	log.Println(dockerArgs)

	// Exec docker (POSIX only)
	err = syscall.Exec(dockerArgs[0], dockerArgs, os.Environ())
	return fmt.Errorf("while invoking docker: %s", err)
}

func dockerChild(info *childInfo) error {
	var cfg dockerConfig

	node, ok := util.DefaultChild(info.args, "image")
	if !ok {
		return fmt.Errorf("no image specified")
	}
	cfg.image, ok = node.(string)
	if !ok {
		return fmt.Errorf("child image is not a string")
	}

	argsMap, ok := info.args.(map[interface{}]interface{})
	if ok {
		var cfgFields = []struct {
			name string
			val  *string
		}{
			{"image", &cfg.image},
			{"user", &cfg.user},
			{"dir", &cfg.dir},
			{"proj", &cfg.projPath},
			{"config", &cfg.configFilename},
		}

		for _, fld := range cfgFields {
			arg, ok := argsMap[fld.name]
			if ok {
				argStr, ok := arg.(string)
				if ok {
					*fld.val = argStr
				} else {
					return fmt.Errorf("%s should be a string; got %q", fld.name, arg)
				}
			}
		}

		arg, ok := argsMap["args"]
		if ok {
			log.Printf("%T %#v", arg, arg)
			argsList, ok := arg.([]interface{})
			if !ok {
				return fmt.Errorf("'args' must be a list")
			}
			for _, argIface := range argsList {
				arg, ok := argIface.(string)
				if !ok {
					return fmt.Errorf("each 'arg' must be a tsring")
				}
				cfg.args = append(cfg.args, arg)
			}
		}
	}

	if cfg.image == "" {
		return fmt.Errorf("no image specified")
	}

	return run(&cfg, info)
}

func init() {
	childFuncs["docker"] = dockerChild
}
