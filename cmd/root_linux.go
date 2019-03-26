package cmd

import (
	"code.cloudfoundry.org/cfdev/config"
	"code.cloudfoundry.org/cfdev/daemon"
	"code.cloudfoundry.org/cfdev/driver"
)

func newDaemonRunner(config config.Config) driver.DaemonRunner {
	return daemon.NewServiceWrapper(config)
}