package stop

import (
	"code.cloudfoundry.org/cfdev/config"
	"code.cloudfoundry.org/cfdev/process"
	"github.com/spf13/cobra"
)

//go:generate mockgen -package mocks -destination mocks/cfdevd_client.go code.cloudfoundry.org/cfdev/cmd/stop CfdevdClient
type CfdevdClient interface {
	Uninstall() (string, error)
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

//go:generate mockgen -package mocks -destination mocks/linuxkit.go code.cloudfoundry.org/cfdev/cmd/stop LinuxKit
type LinuxKit interface {
	Stop() error
	Destroy() error
}

//go:generate mockgen -package mocks -destination mocks/vpnkit.go code.cloudfoundry.org/cfdev/cmd/stop VpnKit
type VpnKit interface {
	Stop() error
	Destroy() error
}

type Stop struct {
	LinuxKit     LinuxKit
	VpnKit       VpnKit
	Config       config.Config
	HyperV       *process.HyperV
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
