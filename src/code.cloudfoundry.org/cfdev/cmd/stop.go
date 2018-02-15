package cmd

import (
	"path/filepath"
	"code.cloudfoundry.org/cfdev/user"
	"code.cloudfoundry.org/cfdev/process"
	"fmt"
)

type Stop struct{}

func(s *Stop) Run(args []string) error {
	devHome, _ := user.CFDevHome()
	linuxkitPid := filepath.Join(devHome, "state", "linuxkit.pid")
	vpnkitPid := filepath.Join(devHome, "state", "vpnkit.pid")
	hyperkitPid := filepath.Join(devHome, "state", "hyperkit.pid")

	if err := process.Terminate(linuxkitPid); err != nil {
		return fmt.Errorf("failed to terminate linuxkit: %s", err)
	}
	if err := process.Terminate(vpnkitPid); err != nil {
		return fmt.Errorf("failed to terminate vpnkit: %s", err)
	}
	if err := process.Kill(hyperkitPid); err != nil {
		return fmt.Errorf("failed to terminate hyperkit: %s", err)
	}
	return nil
}
