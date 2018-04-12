package env

import (
	"fmt"
	"os"
	"strings"

	"io/ioutil"
	"path"

	"code.cloudfoundry.org/cfdev/config"
)

type ProxyConfig struct {
	Http    string `json:"http,omitempty"`
	Https   string `json:"https,omitempty"`
	NoProxy string `json:"exclude,omitempty"`
}

func BuildProxyConfig(boshDirectorIp string, cfRouterIp string) ProxyConfig {
	httpProxy := os.Getenv("http_proxy")
	if os.Getenv("HTTP_PROXY") != "" {
		httpProxy = os.Getenv("HTTP_PROXY")
	}

	httpsProxy := os.Getenv("https_proxy")
	if os.Getenv("HTTPS_PROXY") != "" {
		httpsProxy = os.Getenv("HTTPS_PROXY")
	}

	noProxy := os.Getenv("no_proxy")
	if os.Getenv("NO_PROXY") != "" {
		noProxy = os.Getenv("NO_PROXY")
	}

	if boshDirectorIp != "" && !strings.Contains(noProxy, boshDirectorIp) {
		noProxy = strings.Join([]string{noProxy, boshDirectorIp}, ",")
	}

	if cfRouterIp != "" && !strings.Contains(noProxy, cfRouterIp) {
		noProxy = strings.Join([]string{noProxy, cfRouterIp}, ",")
	}

	proxyConfig := ProxyConfig{
		Http:    httpProxy,
		Https:   httpsProxy,
		NoProxy: noProxy,
	}

	return proxyConfig
}

func Setup(config config.Config) error {
	if err := os.MkdirAll(config.CFDevHome, 0755); err != nil {
		return fmt.Errorf("failed to create home dir at path %s: %s", config.CFDevHome, err)
	}

	if err := os.MkdirAll(config.CacheDir, 0755); err != nil {
		return fmt.Errorf("failed to create cache dir at path %s: %s", config.CacheDir, err)
	}

	if err := os.MkdirAll(config.StateDir, 0755); err != nil {
		return fmt.Errorf("failed to create state dir at path %s: %s", config.StateDir, err)
	}

	return nil
}

func SetupAnalytics(config config.Config) error {

	if err := os.MkdirAll(config.AnalyticsDir, 0755); err != nil {
		return fmt.Errorf("failed to create analytics dir at path %s: %s", config.AnalyticsDir, err)
	}

	analyticsFilePath := path.Join(config.AnalyticsDir, config.AnalyticsFile)

	if _, err := os.Stat(analyticsFilePath); err == nil {
		return nil
	}

	if err := ioutil.WriteFile(analyticsFilePath, []byte(""), 0755); err != nil {
		return fmt.Errorf("failed to create analytics text file at path %s: %s", analyticsFilePath, err)
	}

	return nil
}
