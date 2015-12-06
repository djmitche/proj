package child

import (
	"fmt"
	"github.com/djmitche/proj/proj/util"
	"log"
	"os"
	"os/exec"
	"syscall"
)

func sshChild(info *childInfo) error {
	var host string
	var configFilename string
	var user string
	var projPath = "proj"

	node, ok := util.DefaultChild(info.args, "host")
	if !ok {
		return fmt.Errorf("no host specified")
	}
	host, ok = node.(string)
	if !ok {
		return fmt.Errorf("child host is not a string")
	}

	argsMap, ok := info.args.(map[interface{}]interface{})
	if ok {
		configArg, ok := argsMap["config"]
		if ok {
			configArgStr, ok := configArg.(string)
			if ok {
				configFilename = configArgStr
			} else {
				return fmt.Errorf("config should be a string")
			}
		}
		userArg, ok := argsMap["user"]
		if ok {
			userArgStr, ok := userArg.(string)
			if ok {
				user = userArgStr
			} else {
				return fmt.Errorf("user should be a string")
			}
		}
		projArg, ok := argsMap["proj"]
		if ok {
			projArgStr, ok := projArg.(string)
			if ok {
				projPath = projArgStr
			} else {
				return fmt.Errorf("proj should be a string")
			}
		}
	}

	return connectBySsh(user, host, configFilename, projPath, info)
}

// XXX also used from ec2Child
func connectBySsh(user, host, configFilename, projPath string, info *childInfo) error {
	log.Printf("Connecting to %q via SSH", host)

	// build an ssh command line
	var sshArgs []string
	sshPath, err := exec.LookPath("ssh")
	if err != nil {
		return fmt.Errorf("'ssh' not found: %s", err)
	}

	sshArgs = append(sshArgs, sshPath)
	sshArgs = append(sshArgs, "-t")
	if user != "" {
		sshArgs = append(sshArgs, "-l")
		sshArgs = append(sshArgs, user)
	}
	sshArgs = append(sshArgs, host)
	sshArgs = append(sshArgs, projPath)
	if configFilename != "" {
		sshArgs = append(sshArgs, "-config")
		sshArgs = append(sshArgs, configFilename)
	}

	// TODO: support running proj in a subdir on the remote system

	// TODO: support setting ForwardAgent

	// add the path, quoting it for ssh if necessary
	if info.path == "" {
		sshArgs = append(sshArgs, "''")
	} else {
		sshArgs = append(sshArgs, info.path)
	}

	// Exec SSH (POSIX only)
	err = syscall.Exec(sshArgs[0], sshArgs, os.Environ())
	return fmt.Errorf("while invoking ssh: %s", host, err)
}

func init() {
	childFuncs["ssh"] = sshChild
}
