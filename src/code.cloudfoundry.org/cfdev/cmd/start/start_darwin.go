package start

import (
	"code.cloudfoundry.org/cfdev/errors"
)

func (s *Start) startVM(args Args, depsIsoPath string) error {
	if err := s.LinuxKit.Start(args.Cpus, args.Mem, depsIsoPath); err != nil {
		return err
	}
	return nil
}

func (s *Start) osSpecificSetup(args Args, depsIsoPath string) error {
	s.UI.Say("Installing cfdevd network helper...")
	if err := s.CFDevD.Install(); err != nil {
		return errors.SafeWrap(err, "installing cfdevd")
	}
	return nil
}

func (s *Start) abort() {
	s.LinuxKit.Stop()
	s.VpnKit.Stop()
}

func (s *Start) isRunning() (bool, error) {
	return s.LinuxKit.IsRunning()
}
