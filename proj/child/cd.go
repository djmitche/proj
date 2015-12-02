package child

import (
	"fmt"
	"github.com/djmitche/proj/proj/util"
	"os"
)

func cdChild(info *childInfo) error {
	var dir string
	var configFilename string

	node, ok := util.DefaultChild(info.args, "dir")
	if !ok {
		return fmt.Errorf("no dir specified")
	}
	dir, ok = node.(string)
	if !ok {
		return fmt.Errorf("child dir is not a string")
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
	}

	// actually start the child
	err := os.Chdir(dir)
	if err != nil {
		return err
	}

	// re-run proj from the top, in the same process
	return info.recurse(info.context, configFilename, info.path)
}

func init() {
	childFuncs["cd"] = cdChild
}
