package stop

import (
	"github.com/spf13/cobra"
	"code.cloudfoundry.org/cfdev/cfanalytics"
	"code.cloudfoundry.org/cfdev/process"
	"path/filepath"
	"code.cloudfoundry.org/cfdevd/launchd"
	"code.cloudfoundry.org/cfdev/errors"
)



func (s *Stop) RunE(cmd *cobra.Command, args []string) error {
	s.Analytics.Event(cfanalytics.STOP)

	var reterr error

	if err := s.Launchd.RemoveDaemon(daemonSpec(process.LinuxKitLabel)); err != nil {
		reterr = errors.SafeWrap(err, "failed to stop linuxkit")
	}

	if err := s.Launchd.RemoveDaemon(daemonSpec(process.VpnKitLabel)); err != nil {
		reterr = errors.SafeWrap(err, "failed to stop vpnkit")
	}

	if err := s.ProcManager.SafeKill(filepath.Join(s.Config.StateDir, "hyperkit.pid"), "hyperkit"); err != nil {
		reterr = errors.SafeWrap(err, "failed to kill hyperkit")
	}

	if _, err := s.CfdevdClient.Uninstall(); err != nil {
		reterr = errors.SafeWrap(err, "failed to uninstall cfdevd")
	}

	if reterr != nil {
		return errors.SafeWrap(reterr, "cf dev stop")
	}
	return nil
}

func daemonSpec(label string) launchd.DaemonSpec {
	return launchd.DaemonSpec{
		Label: label,
	}
}
