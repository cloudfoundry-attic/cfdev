package cmd

import (
	"code.cloudfoundry.org/cfdev/config"
	"code.cloudfoundry.org/cfdev/daemon"
	"code.cloudfoundry.org/cfdev/driver"
	"code.cloudfoundry.org/cfdev/driver/hyperv"
	"code.cloudfoundry.org/cfdev/host"
	"code.cloudfoundry.org/cfdev/runner"
)

func newHost() *host.Host {
	return &host.Host{
		Powershell: &runner.Powershell{},
	}
}

func newDaemonRunner(config config.Config) driver.DaemonRunner {
	return daemon.NewWinSW(config.CFDevHome)
}

func newDriver(ui UI, config config.Config) driver.Driver {
	daemonRunner := newDaemonRunner(config)

	return hyperv.New(
		config,
		daemonRunner,
		ui,
		runner.Powershell{},
		"7207f451-2ca3-4b88-8d01-820a21d78293",
		"cc2a519a-fb40-4e45-a9f1-c7f04c5ad7fa",
		"e3ae8f06-8c25-47fb-b6ed-c20702bcef5e",
	)
}
