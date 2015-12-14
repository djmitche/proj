package child

import (
	"fmt"
	"github.com/djmitche/proj/proj/ssh"
	"github.com/djmitche/proj/proj/util"
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

	return ssh.Run(&ssh.Config{
		User:           user,
		Host:           host,
		ForwardAgent:   true,
		ConfigFilename: configFilename,
		ProjPath:       projPath,
		Path:           info.path,
	})
}

func init() {
	childFuncs["ssh"] = sshChild
}
