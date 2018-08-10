package stop

import (
	"code.cloudfoundry.org/cfdev/cfanalytics"
	"code.cloudfoundry.org/cfdev/errors"
	"github.com/spf13/cobra"
)

func (s *Stop) RunE(cmd *cobra.Command, args []string) error {
	s.Analytics.Event(cfanalytics.STOP)

	var reterr error

	if err := s.LinuxKit.Stop(); err != nil {
		reterr = errors.SafeWrap(err, "failed to stop linuxkit")
	}

	if err := s.LinuxKit.Destroy(); err != nil {
		reterr = errors.SafeWrap(err, "failed to destroy linuxkit")
	}

	if err := s.VpnKit.Stop(); err != nil {
		reterr = errors.SafeWrap(err, "failed to stop vpnkit")
	}

	if err := s.VpnKit.Destroy(); err != nil {
		reterr = errors.SafeWrap(err, "failed to destroy vpnkit")
	}

	if _, err := s.CfdevdClient.Uninstall(); err != nil {
		reterr = errors.SafeWrap(err, "failed to uninstall cfdevd")
	}

	if err := s.HostNet.RemoveLoopbackAliases(s.Config.BoshDirectorIP, s.Config.CFRouterIP); err != nil {
		reterr = errors.SafeWrap(err, "failed to remove IP aliases")
	}

	if reterr != nil {
		return errors.SafeWrap(reterr, "cf dev stop")
	}
	return nil
}