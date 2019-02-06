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
	environmentVariables := map[string]string{
		"CFDEV_MODE":         os.Getenv("CFDEV_MODE"),
		"CFDEV_BEHIND_PROXY": fmt.Sprintf("%t", env.IsBehindProxy()),
	}

	proxyConf := env.BuildProxyConfig(
		a.Config.BoshDirectorIP,
		a.Config.CFRouterIP,
		a.Config.HostIP,
	)

	if proxyConf.Http != "" {
		environmentVariables["HTTP_PROXY"] = proxyConf.Http
	}
	if proxyConf.Https != "" {
		environmentVariables["HTTPS_PROXY"] = proxyConf.Https
	}
	if proxyConf.NoProxy != "" {
		environmentVariables["NO_PROXY"] = proxyConf.NoProxy
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
