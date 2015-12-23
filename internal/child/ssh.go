package child

import (
	"fmt"
	"github.com/djmitche/proj/internal/ssh"
)

func sshChild(info *childInfo) error {
	host := info.childConfig.Ssh.Host
	sshHostConfig, ok := info.hostConfig.Ssh[host]
	if !ok {
		return fmt.Errorf("no ssh configuration for host %q", host)
	}

	return ssh.Run(sshHostConfig.Hostname, &sshHostConfig.SshCommonConfig, info.path)
}

func init() {
	childFuncs["ssh"] = sshChild
}
