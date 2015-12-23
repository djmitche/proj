package child

import (
	"os"
)

func cdChild(info *childInfo) error {
	err := os.Chdir(info.childConfig.Cd.Dir)
	if err != nil {
		return err
	}

	// re-run proj from the top, in the same process
	return info.recurse(info.path)
}

func init() {
	childFuncs["cd"] = cdChild
}
