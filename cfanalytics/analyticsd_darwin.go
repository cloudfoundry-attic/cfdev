package cfanalytics

import (
	"code.cloudfoundry.org/cfdev/daemon"
	"os"
	"path"
	"path/filepath"
)

func (a *AnalyticsD) DaemonSpec() daemon.DaemonSpec {
	var (
		proxyConfig          = a.Config.BuildProxyConfig()
		environmentVariables = map[string]string{
			"CFDEV_MODE": os.Getenv("CFDEV_MODE"),
		}
	)

	if proxyConfig.Http != "" {
		environmentVariables["HTTP_PROXY"] = proxyConfig.Http
	}
	if proxyConfig.Https != "" {
		environmentVariables["HTTPS_PROXY"] = proxyConfig.Https
	}
	if proxyConfig.NoProxy != "" {
		environmentVariables["NO_PROXY"] = proxyConfig.NoProxy
	}

	return daemon.DaemonSpec{
		Label:                AnalyticsDLabel,
		Program:              filepath.Join(a.Config.CacheDir, "analyticsd"),
		SessionType:          "Background",
		ProgramArguments:     []string{filepath.Join(a.Config.CacheDir, "analyticsd")},
		EnvironmentVariables: environmentVariables,
		RunAtLoad:            false,
		StdoutPath:           path.Join(a.Config.LogDir, "analyticsd.stdout.log"),
		StderrPath:           path.Join(a.Config.LogDir, "analyticsd.stderr.log"),
	}
}
