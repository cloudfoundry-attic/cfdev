package config

import (
	"os"
	"strings"
)

type ProxyConfig struct {
	Http    string `json:"http,omitempty"`
	Https   string `json:"https,omitempty"`
	NoProxy string `json:"exclude,omitempty"`
}

func (c *Config) BuildProxyConfig() ProxyConfig {
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

	if c.BoshDirectorIP != "" && !strings.Contains(noProxy, c.BoshDirectorIP) {
		noProxy = strings.Join([]string{noProxy, c.BoshDirectorIP}, ",")
	}

	if c.CFRouterIP != "" && !strings.Contains(noProxy, c.CFRouterIP) {
		noProxy = strings.Join([]string{noProxy, c.CFRouterIP}, ",")
	}

	if c.HostIP != "" && !strings.Contains(noProxy, c.HostIP) {
		noProxy = strings.Join([]string{noProxy, c.HostIP}, ",")
	}

	return ProxyConfig{
		Http:    httpProxy,
		Https:   httpsProxy,
		NoProxy: noProxy,
	}
}

func IsBehindProxy() bool {
	return os.Getenv("HTTP_PROXY") != "" ||
		os.Getenv("http_proxy") != "" ||
		os.Getenv("HTTPS_PROXY") != "" ||
		os.Getenv("https_proxy") != ""
}
