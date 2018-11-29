package stop

import (
	"fmt"
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

//go:generate mockgen -package mocks -destination mocks/host.go code.cloudfoundry.org/cfdev/cmd/stop Host
type Host interface {
	CheckRequirements() error
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

//go:generate mockgen -package mocks -destination mocks/analyticsd.go code.cloudfoundry.org/cfdev/cmd/stop AnalyticsD
type AnalyticsD interface {
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
	AnalyticsD   AnalyticsD
	Host         Host
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

	if err := s.Host.CheckRequirements(); err != nil {
		return err
	}

	var reterr error

	fmt.Printf("DEBUG: STOP: ABOUT TO STOP analytics \n")
	if err := s.AnalyticsD.Stop(); err != nil {
		reterr = errors.SafeWrap(err, "failed to stop analyticsd")
	}

	if err := s.AnalyticsD.Destroy(); err != nil {
		reterr = errors.SafeWrap(err, "failed to destroy analyticsd")
	}

	fmt.Printf("DEBUG: STOP: ABOUT TO STOP HYPERV \n")
	if err := s.Hypervisor.Stop(vmName); err != nil {
		reterr = errors.SafeWrap(err, "failed to stop the VM")
	}

	if err := s.Hypervisor.Destroy(vmName); err != nil {
		reterr = errors.SafeWrap(err, "failed to destroy the VM")
	}

	fmt.Printf("DEBUG: STOP: ABOUT TO STOP VPNKIT \n")
	if err := s.VpnKit.Stop(); err != nil {
		reterr = errors.SafeWrap(err, "failed to stop vpnkit")
	}

	if err := s.VpnKit.Destroy(); err != nil {
		reterr = errors.SafeWrap(err, "failed to destroy vpnkit")
	}

	fmt.Printf("DEBUG: STOP: REALLY SHOULD HAVE STOPPED VPNKIT \n")

	if err := s.HostNet.RemoveLoopbackAliases(s.Config.BoshDirectorIP, s.Config.CFRouterIP); err != nil {
		reterr = errors.SafeWrap(err, "failed to remove IP aliases")
	}

	if runtime.GOOS == "darwin" {
		if _, err := s.CfdevdClient.Uninstall(); err != nil {
			reterr = errors.SafeWrap(err, "failed to uninstall cfdevd")
		}
	}

	if reterr != nil {
		return errors.SafeWrap(reterr, "cf dev stop")
	}
	return nil
}
