package hypervisor

import (
	"code.cloudfoundry.org/cfdev/daemon"
)

func (l *LinuxKit) DaemonSpec(cpus, mem int) (daemon.DaemonSpec, error) {
	return daemon.DaemonSpec{}, nil
}

