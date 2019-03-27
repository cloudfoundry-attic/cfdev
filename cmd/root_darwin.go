package cmd

import (
	"code.cloudfoundry.org/cfdev/config"
	"code.cloudfoundry.org/cfdev/daemon"
	"code.cloudfoundry.org/cfdev/driver"
	"code.cloudfoundry.org/cfdev/driver/hyperkit"
	cfdevdClient "code.cloudfoundry.org/cfdev/pkg/cfdevd/client"
)

func newDaemonRunner(config config.Config) driver.DaemonRunner {
	return daemon.New(config.StateDir)
}

func newDriver(ui UI, config config.Config) driver.Driver {
	var (
		daemonRunner = newDaemonRunner(config)
		cfdevd       = cfdevdClient.New("CFD3V", config.CFDevDSocketPath)
	)

	return hyperkit.New(config, daemonRunner, ui, cfdevd)
}
