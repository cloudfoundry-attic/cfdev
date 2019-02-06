package cfanalytics

import (
	"code.cloudfoundry.org/cfdev/daemon"
	"os"
	"path"
	"path/filepath"
)

func (a *AnalyticsD) DaemonSpec() daemon.DaemonSpec {
	return daemon.DaemonSpec{
		Label:            AnalyticsDLabel,
		Program:          filepath.Join(a.Config.CacheDir, "analyticsd"),
		SessionType:      "Background",
		ProgramArguments: []string{filepath.Join(a.Config.CacheDir, "analyticsd")},
		EnvironmentVariables: map[string]string{
			"CFDEV_MODE": os.Getenv("CFDEV_MODE"),
		},
		RunAtLoad:        false,
		StdoutPath:       path.Join(a.Config.LogDir, "analyticsd.stdout.log"),
		StderrPath:       path.Join(a.Config.LogDir, "analyticsd.stderr.log"),
	}
}
