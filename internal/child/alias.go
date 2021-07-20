package child

import "fmt"

func aliasChild(info *childInfo) error {
	// re-run proj from the top, in the same process
	return info.recurse(fmt.Sprintf("%s/%s", info.childConfig.Alias.Target, info.path))
}

func init() {
	childFuncs["alias"] = aliasChild
}
