package cmd

import (
	"path/filepath"
	"code.cloudfoundry.org/cfdev/user"
	"code.cloudfoundry.org/cfdev/process"
)

type Stop struct{}

func(s *Stop) Run(args []string) {
	devHome, _ := user.CFDevHome()
	linuxkitPid := filepath.Join(devHome, "state", "linuxkit.pid")
	vpnkitPid := filepath.Join(devHome, "state", "vpnkit.pid")
	hyperkitPid := filepath.Join(devHome, "state", "hyperkit.pid")

	process.Terminate(linuxkitPid)
	process.Terminate(vpnkitPid)
	process.Kill(hyperkitPid)
}
