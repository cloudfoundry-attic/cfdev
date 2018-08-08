package start

import (
	"code.cloudfoundry.org/cfdev/errors"
	"code.cloudfoundry.org/cfdev/process"
)

func (s *Start) osSpecificSetup(args Args, depsIsoPath string) error {
	s.UI.Say("Creating the VM...")
	if err := s.HyperV.CreateVM(process.VM{
		DepsIso:  depsIsoPath,
		MemoryMB: args.Mem,
		CPUs:     args.Cpus,
	}); err != nil {
		return errors.SafeWrap(err, "Unable to create VM")
	}
	return nil
}

func (s *Start) startVM(args Args, depsIsoPath string) error {
	return s.HyperV.Start("cfdev")
}

func (s *Start) abort() {
	s.LinuxKit.Stop()
	s.VpnKit.Stop()
}

func (s *Start) isRunning() (bool, error) {
	return false, nil
}
