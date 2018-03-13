package cmd

import (
	"fmt"
	"sync"
	"syscall"

	"code.cloudfoundry.org/cfdev/config"
	"code.cloudfoundry.org/cfdev/process"
)

type Stop struct {
	Config config.Config
}

func (s *Stop) Run(args []string) error {
	var reterr error
	var all sync.WaitGroup
	all.Add(3)

	go func() {
		defer all.Done()
		if err := process.SignalAndCleanup(s.Config.LinuxkitPidFile, s.Config.CFDevHome, syscall.SIGTERM); err != nil {
			reterr = fmt.Errorf("failed to terminate linuxkit: %s", err)
		}
	}()
	go func() {
		defer all.Done()
		if err := process.SignalAndCleanup(s.Config.VpnkitPidFile, s.Config.CFDevHome, syscall.SIGTERM); err != nil {
			reterr = fmt.Errorf("failed to terminate vpnkit: %s", err)
		}
	}()
	go func() {
		defer all.Done()
		if err := process.SignalAndCleanup(s.Config.HyperkitPidFile, s.Config.CFDevHome, syscall.SIGKILL); err != nil {
			reterr = fmt.Errorf("failed to terminate hyperkit: %s", err)
		}
	}()

	all.Wait()

	return reterr
}
