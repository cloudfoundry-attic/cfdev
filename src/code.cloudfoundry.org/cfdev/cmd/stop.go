package cmd

import (
	"code.cloudfoundry.org/cfdev/process"
	"fmt"
	"code.cloudfoundry.org/cfdev/config"
)

type Stop struct{
	Config config.Config
}

func(s *Stop) Run(args []string) error {
	if err := process.Terminate(s.Config.LinuxkitPidFile); err != nil {
		return fmt.Errorf("failed to terminate linuxkit: %s", err)
	}
	if err := process.Terminate(s.Config.VpnkitPidFile); err != nil {
		return fmt.Errorf("failed to terminate vpnkit: %s", err)
	}
	if err := process.Kill(s.Config.HyperkitPidFile); err != nil {
		return fmt.Errorf("failed to terminate hyperkit: %s", err)
	}
	return nil
}
