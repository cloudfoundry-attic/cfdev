package env

import (
	"code.cloudfoundry.org/cfdev/resource"
	"fmt"
	"os"
	"runtime"
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
	err := e.RemoveDirAlls(
		e.Config.LogDir,
		e.Config.ServicesDir,
		e.Config.StateDir)
	if err != nil {
		return err
	}

	return e.MkdirAlls(
		e.Config.CFDevHome,
		e.Config.CacheDir,
		e.Config.VpnKitStateDir,
		e.Config.StateLinuxkit,
		e.Config.StateBosh,
		e.Config.ServicesDir,
		e.Config.LogDir)
}

func (e *Env) MkdirAlls(dirs ...string) error {
	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return errors.SafeWrap(fmt.Errorf("path %s: %s", dir, err), "failed to create dir")
		}
	}

	return nil
}

func (e *Env) RemoveDirAlls(dirs ...string) error {
	for _, dir := range dirs {
		if err := os.RemoveAll(dir); err != nil {
			return errors.SafeWrap(fmt.Errorf("path %s: %s", dir, err), "failed to remove dir")
		}
	}

	return nil
}

func (e *Env) SetupState() error {
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
			Include: "id_rsa",
			Dst:     e.Config.CacheDir,
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
	}

	if runtime.GOOS == "windows" {
		thingsToUntar = append(thingsToUntar, resource.TarOpts{
			Include: "disk.vhdx",
				Dst:     e.Config.StateLinuxkit,
			})
	} else {
		thingsToUntar = append(thingsToUntar, resource.TarOpts{
			Include: "disk.qcow2",
				Dst:     e.Config.StateLinuxkit,
			})
	}

	err := resource.Untar(*e.Config.DepsFile, thingsToUntar)
	if err != nil {
		return errors.SafeWrap(err, "failed to untar the desired parts of the tarball")
	}

	return nil
}
