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
		ProgramArguments: []string{filepath.Join(a.Config.CacheDir, "analyticsd"), os.Getenv("CFDEV_MODE")},
		RunAtLoad:        false,
		StdoutPath:       path.Join(a.Config.CFDevHome, "analyticsd.stdout.log"),
		StderrPath:       path.Join(a.Config.CFDevHome, "analyticsd.stderr.log"),
	}
}