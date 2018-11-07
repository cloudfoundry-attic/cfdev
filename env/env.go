package env

import (
	"code.cloudfoundry.org/cfdev/resource"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"code.cloudfoundry.org/cfdev/config"
	"code.cloudfoundry.org/cfdev/errors"
)

type ProxyConfig struct {
	Http    string `json:"http,omitempty"`
	Https   string `json:"https,omitempty"`
	NoProxy string `json:"exclude,omitempty"`
}

func BuildProxyConfig(boshDirectorIP string, cfRouterIP string, hostIP string) ProxyConfig {
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

	if boshDirectorIP != "" && !strings.Contains(noProxy, boshDirectorIP) {
		noProxy = strings.Join([]string{noProxy, boshDirectorIP}, ",")
	}

	if cfRouterIP != "" && !strings.Contains(noProxy, cfRouterIP) {
		noProxy = strings.Join([]string{noProxy, cfRouterIP}, ",")
	}

	if hostIP != "" && !strings.Contains(noProxy, hostIP) {
		noProxy = strings.Join([]string{noProxy, hostIP}, ",")
	}

	proxyConfig := ProxyConfig{
		Http:    httpProxy,
		Https:   httpsProxy,
		NoProxy: noProxy,
	}

	return proxyConfig
}

type Env struct {
	Config config.Config
}

func (e *Env) CreateDirs() error {
	if err := os.MkdirAll(e.Config.CFDevHome, 0755); err != nil {
		return errors.SafeWrap(fmt.Errorf("path %s: %s", e.Config.CFDevHome, err), "failed to create cfdev home dir")
	}

	if err := os.MkdirAll(e.Config.CacheDir, 0755); err != nil {
		return errors.SafeWrap(fmt.Errorf("path %s: %s", e.Config.CacheDir, err), "failed to create cache dir")
	}

	if err := os.MkdirAll(e.Config.VpnKitStateDir, 0755); err != nil {
		return errors.SafeWrap(fmt.Errorf("path %s: %s", e.Config.VpnKitStateDir, err), "failed to create state dir")
	}

	if err := os.MkdirAll(filepath.Join(e.Config.StateLinuxkit), 0755); err != nil {
		return errors.SafeWrap(fmt.Errorf("path %s: %s", filepath.Join(e.Config.StateLinuxkit), err), "failed to create state dir")
	}

	if err := os.MkdirAll(filepath.Join(e.Config.StateBosh), 0755); err != nil {
		return errors.SafeWrap(fmt.Errorf("path %s: %s", filepath.Join(e.Config.StateBosh), err), "failed to create state dir")
	}

	if err := os.MkdirAll(filepath.Join(e.Config.ServicesDir, "logs"), 0755); err != nil {
		return errors.SafeWrap(fmt.Errorf("path %s: %s", filepath.Join(e.Config.ServicesDir, "logs"), err), "failed to create services dir")
	}

	if err := os.MkdirAll(e.Config.LogDir, 0755); err != nil {
		return errors.SafeWrap(fmt.Errorf("path %s: %s", e.Config.LogDir, err), "failed to create log dir")
	}

	return nil
}

func (e *Env) SetupState() error {
	tarFilepath := filepath.Join(e.Config.CacheDir, "cfdev-deps.tgz")

	thingsToUntar := []resource.TarOpts{
		{
			Include: "state.json",
			Dst:     e.Config.StateBosh,
		},
		{
			Include: "creds.yml",
			Dst:     e.Config.StateBosh,
		},
		{
			Include: "secret",
			Dst:     e.Config.StateBosh,
		},
		{
			Include: "jumpbox.key",
			Dst:     e.Config.StateBosh,
		},
		{
			Include: "ca.crt",
			Dst:     e.Config.StateBosh,
		},
		{
			IncludeFolder: "services",
			Dst:           e.Config.CFDevHome,
		},
		{
			IncludeFolder: "binaries",
			FlattenFolder: true,
			Dst:           e.Config.CacheDir,
		},
		{
			IncludeFolder: "deployment_config",
			FlattenFolder: true,
			Dst:           e.Config.CacheDir,
		},
		{
			Include: "disk.qcow2",
			Dst:     e.Config.StateLinuxkit,
		},
	}

	err := resource.Untar(tarFilepath, thingsToUntar)
	if err != nil {
		return errors.SafeWrap(err, "failed to untar the desired parts of the tarball")
	}

	return nil
}
