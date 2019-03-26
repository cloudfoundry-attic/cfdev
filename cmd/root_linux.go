package cmd

import (
	"code.cloudfoundry.org/cfdev/config"
	"code.cloudfoundry.org/cfdev/daemon"
	"code.cloudfoundry.org/cfdev/driver"
	"code.cloudfoundry.org/cfdev/driver/kvm"
	"code.cloudfoundry.org/cfdev/host"
)

func newHost() *host.Host {
	return &host.Host{}
}

func newDaemonRunner(config config.Config) driver.DaemonRunner {
	return daemon.NewServiceWrapper(config)
}

func newDriver(ui UI, config config.Config) driver.Driver {
	daemonRunner := newDaemonRunner(config)

	return kvm.New(config, daemonRunner, ui)
}