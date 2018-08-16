package stop

import (
	"runtime"

	"code.cloudfoundry.org/cfdev/cfanalytics"
	"code.cloudfoundry.org/cfdev/config"
	"code.cloudfoundry.org/cfdev/errors"
	"github.com/spf13/cobra"
)

//go:generate mockgen -package mocks -destination mocks/cfdevd_client.go code.cloudfoundry.org/cfdev/cmd/stop CfdevdClient
type CfdevdClient interface {
	Uninstall() (string, error)
	RemoveIPAlias() (string, error)
}

type UI interface {
	Say(message string, args ...interface{})
}

//go:generate mockgen -package mocks -destination mocks/analytics.go code.cloudfoundry.org/cfdev/cmd/stop Analytics
type Analytics interface {
	Event(event string, data ...map[string]interface{}) error
}

//go:generate mockgen -package mocks -destination mocks/network.go code.cloudfoundry.org/cfdev/cmd/stop HostNet
type HostNet interface {
	RemoveLoopbackAliases(...string) error
}

//go:generate mockgen -package mocks -destination mocks/linuxkit.go code.cloudfoundry.org/cfdev/cmd/stop Hypervisor
type Hypervisor interface {
	Stop(vmName string) error
	Destroy(vmName string) error
}

//go:generate mockgen -package mocks -destination mocks/vpnkit.go code.cloudfoundry.org/cfdev/cmd/stop VpnKit
type VpnKit interface {
	Stop() error
	Destroy() error
}

type Stop struct {
	Hypervisor   Hypervisor
	VpnKit       VpnKit
	Config       config.Config
	CfdevdClient CfdevdClient
	Analytics    Analytics
	HostNet      HostNet
}

func (s *Stop) Cmd() *cobra.Command {
	return &cobra.Command{
		Use:  "stop",
		RunE: s.RunE,
	}
}

const vmName = "cfdev"

func (s *Stop) RunE(cmd *cobra.Command, args []string) error {
	s.Analytics.Event(cfanalytics.STOP)

	var reterr error

	if err := s.Hypervisor.Stop(vmName); err != nil {
		reterr = errors.SafeWrap(err, "failed to stop the VM")
	}

	if err := s.Hypervisor.Destroy(vmName); err != nil {
		reterr = errors.SafeWrap(err, "failed to destroy the VM")
	}

	if err := s.VpnKit.Stop(); err != nil {
		reterr = errors.SafeWrap(err, "failed to stop vpnkit")
	}

	if err := s.VpnKit.Destroy(); err != nil {
		reterr = errors.SafeWrap(err, "failed to destroy vpnkit")
	}

	if runtime.GOOS == "darwin" {
		if _, err := s.CfdevdClient.RemoveIPAlias(); err != nil {
			reterr = errors.SafeWrap(err, "failed to remove IP aliases")
		}

		if _, err := s.CfdevdClient.Uninstall(); err != nil {
			reterr = errors.SafeWrap(err, "failed to uninstall cfdevd")
		}

	} else {
		if err := s.HostNet.RemoveLoopbackAliases(s.Config.BoshDirectorIP, s.Config.CFRouterIP); err != nil {
			reterr = errors.SafeWrap(err, "failed to remove IP aliases")
		}
	}

	if reterr != nil {
		return errors.SafeWrap(reterr, "cf dev stop")
	}
	return nil
}
