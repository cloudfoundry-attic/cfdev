package cmd

import (
	cfdevdClient "code.cloudfoundry.org/cfdev/cfdevd/client"
	"code.cloudfoundry.org/cfdev/config"
	"code.cloudfoundry.org/cfdev/daemon"
	"code.cloudfoundry.org/cfdev/driver"
	"code.cloudfoundry.org/cfdev/driver/hyperkit"
	"code.cloudfoundry.org/cfdev/host"
)

func newHost() *host.Host {
	return &host.Host{}
}

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
