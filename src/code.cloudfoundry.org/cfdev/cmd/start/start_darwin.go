package start

import "code.cloudfoundry.org/cfdev/errors"

func (s *Start) osSpecificSetup() error {
	s.UI.Say("Installing cfdevd network helper...")
	if err := s.CFDevD.Install(); err != nil {
		return errors.SafeWrap(err, "installing cfdevd")
	}
	return nil
}
