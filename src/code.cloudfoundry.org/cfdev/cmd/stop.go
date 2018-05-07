package cmd

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"

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

func NewStop(config config.Config, launchd LaunchdStop, cfdevdClient CfdevdClient) *cobra.Command {
	cmd := &cobra.Command{
		Use: "stop",
		RunE: func(cmd *cobra.Command, args []string) error {
			err := runStop(config, launchd, cfdevdClient)
			if err != nil {
				return errors.SafeWrap(err, "cf dev stop")
			}
			return nil
		},
	}
	return cmd
}

func runStop(config config.Config, launchd LaunchdStop, cfdevdClient CfdevdClient) error {
	config.Analytics.Event(cfanalytics.STOP, map[string]interface{}{"type": "cf"})

	var reterr error

	if err := launchd.Stop(process.LinuxKitLabel); err != nil {
		reterr = errors.SafeWrap(err, "failed to stop linuxkit")
	}

	if err := launchd.Stop(process.VpnKitLabel); err != nil {
		reterr = errors.SafeWrap(err, "failed to stop vpnkit")
	}

	if err := killHyperkit(config); err != nil {
		reterr = errors.SafeWrap(err, "failed to kill hyperkit")
	}

	if _, err := cfdevdClient.Uninstall(); err != nil {
		reterr = errors.SafeWrap(err, "failed to uninstall cfdevd")
	}

	return reterr
}

func killHyperkit(config config.Config) error {
	pidfile := filepath.Join(config.StateDir, "hyperkit.pid")
	data, err := ioutil.ReadFile(pidfile)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return err
	}
	pid, err := strconv.Atoi(string(data))
	if err != nil {
		return err
	}
	process, err := os.FindProcess(pid)
	if err != nil {
		return err
	}
	err = process.Kill()
	if err != nil {
		return err
	}
	os.Remove(pidfile)
	return nil
}
