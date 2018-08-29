package cfanalytics

import (
	"path/filepath"
	"code.cloudfoundry.org/cfdev/daemon"
	"code.cloudfoundry.org/cfdev/config"
	"fmt"
)

const AnalyticsDLabel = "org.cloudfoundry.cfdev.cfanalyticsd"

type AnalyticsD struct {
	Config       config.Config
	DaemonRunner DaemonRunner
}

type DaemonRunner interface {
	AddDaemon(daemon.DaemonSpec) error
	RemoveDaemon(string) error
	Start(string) error
	Stop(string) error
	IsRunning(string) (bool, error)
}

func (a *AnalyticsD) Start() error {
	spec, err := a.DaemonSpec()
	if err != nil {
		return err
	}

	fmt.Println("DAEMON SPEC: ", spec)

	err = a.DaemonRunner.AddDaemon(spec)
	if err != nil {
		return err
	}

	err = a.DaemonRunner.Start(AnalyticsDLabel)
	if err != nil {
		return err
	}

	fmt.Println("DAEMON ANALYTICS SHOULD BE RUNNING???")

	return err
}

func (a *AnalyticsD) Stop() error {
	var reterr error
	if err := a.DaemonRunner.Stop(AnalyticsDLabel); err != nil {
		reterr = err
	}
	return reterr
}

func (a *AnalyticsD) Destroy() error {
	return a.DaemonRunner.RemoveDaemon(AnalyticsDLabel)
}

func (a *AnalyticsD) IsRunning() (bool, error) {
	return a.DaemonRunner.IsRunning(AnalyticsDLabel)
}

func (a *AnalyticsD) DaemonSpec() (daemon.DaemonSpec, error) {
	analyticsD := filepath.Join(a.Config.CacheDir, "analyticsd")

	return daemon.DaemonSpec{
		Label:            AnalyticsDLabel,
		Program:          analyticsD,
		SessionType:      "Background",
		ProgramArguments: []string{"analyticsd"},
		RunAtLoad:        false,
	}, nil
}
