package cmd

import (
	"code.cloudfoundry.org/cfdev/config"
	"code.cloudfoundry.org/cfdev/daemon"
	"code.cloudfoundry.org/cfdev/driver"
	"code.cloudfoundry.org/cfdev/driver/kvm"
)

func newDaemonRunner(config config.Config) driver.DaemonRunner {
	return daemon.NewServiceWrapper(config)
}

func newDriver(ui UI, config config.Config) driver.Driver {
	daemonRunner := newDaemonRunner(config)

	return kvm.New(config, daemonRunner, ui)
}