package cmd

import (
	"code.cloudfoundry.org/cfdev/cfanalytics"
	"code.cloudfoundry.org/cfdev/config"
	"code.cloudfoundry.org/cfdev/errors"
	"code.cloudfoundry.org/cfdev/process"
	"github.com/spf13/cobra"
)

type LaunchdStop interface {
	Stop(label string) error
}

type CfdevdClient interface {
	Uninstall() (string, error)
}

func NewStop(Config config.Config, Launchd LaunchdStop, CfdevdClient CfdevdClient) *cobra.Command {
	cmd := &cobra.Command{
		Use: "stop",
		RunE: func(cmd *cobra.Command, args []string) error {
			err := runStop(Config, Launchd, CfdevdClient)
			if err != nil {
				return errors.SafeWrap(err, "cf dev stop")
			}
			return nil
		},
	}
	return cmd
}

func runStop(Config config.Config, Launchd LaunchdStop, CfdevdClient CfdevdClient) error {
	Config.Analytics.Event(cfanalytics.STOP, map[string]interface{}{"type": "cf"})

	var reterr error

	if err := Launchd.Stop(process.LinuxKitLabel); err != nil {
		reterr = errors.SafeWrap(err, "failed to stop linuxkit")
	}

	if err := Launchd.Stop(process.VpnKitLabel); err != nil {
		reterr = errors.SafeWrap(err, "failed to stop vpnkit")
	}

	if _, err := CfdevdClient.Uninstall(); err != nil {
		reterr = errors.SafeWrap(err, "failed to uninstall cfdevd")
	}

	return reterr
}
