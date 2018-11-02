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

	return nil
}

func SetupState(config config.Config) error {
	tarFilepath := filepath.Join(config.CacheDir, "cfdev-deps.tgz")

	qcowPath := filepath.Join(config.StateLinuxkit, "disk.qcow2")
	if _, err := os.Stat(qcowPath); os.IsNotExist(err) {
		err = resource.Untar(config.StateLinuxkit, tarFilepath, resource.TarOpts{Include: "disk.qcow2"})
		if err != nil {
			errors.SafeWrap(fmt.Errorf("%s", err), "unable to untar disk.qcow2")
			return err
		}
	}

	err := resource.Untar(config.StateBosh, tarFilepath, resource.TarOpts{Include: "state.json"})
	if err != nil {
		errors.SafeWrap(fmt.Errorf("%s", err), "unable to untar state.json")
		return err
	}

	err = resource.Untar(config.StateBosh, tarFilepath, resource.TarOpts{Include: "creds.yml"})
	if err != nil {
		errors.SafeWrap(fmt.Errorf("%s", err), "unable to untar creds.yml")
		return err
	}

	err = resource.Untar(config.StateBosh, tarFilepath, resource.TarOpts{Include: "secret"})
	if err != nil {
		errors.SafeWrap(fmt.Errorf("%s", err), "unable to untar secrets")
		return err
	}

	resource.Untar(config.StateBosh, tarFilepath, resource.TarOpts{Include: "jumpbox.key"})
	if err != nil {
		errors.SafeWrap(fmt.Errorf("%s", err), "unable to untar jumpbox.key")
		return err
	}

	resource.Untar(config.StateBosh, tarFilepath, resource.TarOpts{Include: "ca.crt"})
	if err != nil {
		errors.SafeWrap(fmt.Errorf("%s", err), "unable to untar ca.crt")
		return err
	}

	resource.Untar(config.CFDevHome, tarFilepath, resource.TarOpts{IncludeFolder: "services"})
	if err != nil {
		errors.SafeWrap(fmt.Errorf("%s", err), "unable to untar services")
		return err
	}

	resource.Untar(config.CacheDir, tarFilepath, resource.TarOpts{IncludeFolder: "binaries", FlattenFolder: true})
	if err != nil {
		errors.SafeWrap(fmt.Errorf("%s", err), "unable to untar binaries")
		return err
	}

	resource.Untar(config.CacheDir, tarFilepath, resource.TarOpts{IncludeFolder: "deployment_config", FlattenFolder: true})
	if err != nil {
		errors.SafeWrap(fmt.Errorf("%s", err), "unable to untar deployment configuration files")
		return err
	}

	return nil
}
