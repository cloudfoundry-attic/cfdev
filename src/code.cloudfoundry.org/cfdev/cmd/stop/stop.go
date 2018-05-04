package stop

import (
	"path/filepath"

	"code.cloudfoundry.org/cfdev/cfanalytics"
	"code.cloudfoundry.org/cfdev/config"
	"code.cloudfoundry.org/cfdev/errors"
	"code.cloudfoundry.org/cfdev/process"
	"github.com/spf13/cobra"
)

//go:generate mockgen -package mocks -destination mocks/launchd.go code.cloudfoundry.org/cfdev/cmd/stop Launchd
type Launchd interface {
	RemoveDaemon(label string) error
}

//go:generate mockgen -package mocks -destination mocks/cfdevd_client.go code.cloudfoundry.org/cfdev/cmd/stop CfdevdClient
type CfdevdClient interface {
	Uninstall() (string, error)
}

//go:generate mockgen -package mocks -destination mocks/process_manager.go code.cloudfoundry.org/cfdev/cmd/stop ProcManager
type ProcManager interface {
	SafeKill(string, string) error
}

type UI interface {
	Say(message string, args ...interface{})
}

//go:generate mockgen -package mocks -destination mocks/analytics.go code.cloudfoundry.org/cfdev/cmd/stop Analytics
type Analytics interface {
	Event(event string, data ...map[string]interface{}) error
}

type Stop struct {
	Config       config.Config
	Launchd      Launchd
	ProcManager  ProcManager
	CfdevdClient CfdevdClient
	Analytics    Analytics
}

func (s *Stop) Cmd() *cobra.Command {
	return &cobra.Command{
		Use:  "stop",
		RunE: s.RunE,
	}
}

func (s *Stop) RunE(cmd *cobra.Command, args []string) error {
	s.Analytics.Event(cfanalytics.STOP)

	var reterr error

	if err := s.Launchd.RemoveDaemon(process.LinuxKitLabel); err != nil {
		reterr = errors.SafeWrap(err, "failed to stop linuxkit")
	}

	if err := s.Launchd.RemoveDaemon(process.VpnKitLabel); err != nil {
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
