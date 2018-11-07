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

func CreateDirs(config config.Config) error {
	if err := os.MkdirAll(config.CFDevHome, 0755); err != nil {
		return errors.SafeWrap(fmt.Errorf("path %s: %s", config.CFDevHome, err), "failed to create cfdev home dir")
	}

	if err := os.MkdirAll(config.CacheDir, 0755); err != nil {
		return errors.SafeWrap(fmt.Errorf("path %s: %s", config.CacheDir, err), "failed to create cache dir")
	}

	if err := os.MkdirAll(config.VpnKitStateDir, 0755); err != nil {
		return errors.SafeWrap(fmt.Errorf("path %s: %s", config.VpnKitStateDir, err), "failed to create state dir")
	}

	if err := os.MkdirAll(filepath.Join(config.StateLinuxkit), 0755); err != nil {
		return errors.SafeWrap(fmt.Errorf("path %s: %s", filepath.Join(config.StateLinuxkit), err), "failed to create state dir")
	}

	if err := os.MkdirAll(filepath.Join(config.StateBosh), 0755); err != nil {
		return errors.SafeWrap(fmt.Errorf("path %s: %s", filepath.Join(config.StateBosh), err), "failed to create state dir")
	}

	if err := os.MkdirAll(filepath.Join(config.ServicesDir, "logs"), 0755); err != nil {
		return errors.SafeWrap(fmt.Errorf("path %s: %s", filepath.Join(config.ServicesDir), err), "failed to create services dir")
	}

	return nil
}

func SetupState(config config.Config) error {
	tarFilepath := filepath.Join(config.CacheDir, "cfdev-deps.tgz")

	thingsToUntar := []resource.TarOpts{
		{
			Include: "state.json",
			Dst:     config.StateBosh,
		},
		{
			Include: "creds.yml",
			Dst:     config.StateBosh,
		},
		{
			Include: "secret",
			Dst:     config.StateBosh,
		},
		{
			Include: "jumpbox.key",
			Dst:     config.StateBosh,
		},
		{
			Include: "ca.crt",
			Dst:     config.StateBosh,
		},
		{
			IncludeFolder: "services",
			Dst:           config.CFDevHome,
		},
		{
			IncludeFolder: "binaries",
			FlattenFolder: true,
			Dst:           config.CacheDir,
		},
		{
			IncludeFolder: "deployment_config",
			FlattenFolder: true,
			Dst:           config.CacheDir,
		},
		{
			Include: "disk.qcow2",
			Dst:     config.StateLinuxkit,
		},
	}

	err := resource.Untar(tarFilepath, thingsToUntar)
	if err != nil {
		return errors.SafeWrap(err, "failed to untar the desired parts of the tarball")
	}

	return nil
}
