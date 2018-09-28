package cfanalytics

import (
	"code.cloudfoundry.org/cfdev/daemon"
	"os"
	"path/filepath"
)

func (a *AnalyticsD) DaemonSpec() daemon.DaemonSpec {
	return daemon.DaemonSpec{
		Label:            AnalyticsDLabel,
		Program:          filepath.Join(a.Config.CacheDir, "analyticsd.exe"),
		SessionType:      "Background",
		ProgramArguments: []string{os.Getenv("CFDEV_MODE")},
	}
}