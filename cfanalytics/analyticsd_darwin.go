package cfanalytics

import (
	"code.cloudfoundry.org/cfdev/daemon"
	"code.cloudfoundry.org/cfdev/env"
	"fmt"
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
			"CFDEV_BEHIND_PROXY": fmt.Sprintf("%t", env.IsBehindProxy()),
		},
		RunAtLoad:        false,
		StdoutPath:       path.Join(a.Config.LogDir, "analyticsd.stdout.log"),
		StderrPath:       path.Join(a.Config.LogDir, "analyticsd.stderr.log"),
	}
}
