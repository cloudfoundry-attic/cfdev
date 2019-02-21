package env

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
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

func IsBehindProxy() bool {
	return os.Getenv("HTTP_PROXY") != "" ||
		os.Getenv("http_proxy") != "" ||
		os.Getenv("HTTPS_PROXY") != "" ||
		os.Getenv("https_proxy") != ""
}

type Env struct {
	Config config.Config
}

func (e *Env) CreateDirs() error {
	err := e.removeDirAlls(
		e.Config.LogDir,
		e.Config.StateDir,
		e.Config.BinaryDir,
		e.Config.ServicesDir,
		e.Config.DaemonDir)
	if err != nil {
		return err
	}

	return e.mkdirAlls(
		e.Config.CacheDir,
		e.Config.DaemonDir,
		e.Config.LogDir)
}

func (e *Env) mkdirAlls(dirs ...string) error {
	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return errors.SafeWrap(fmt.Errorf("path %s: %s", dir, err), "failed to create dir")
		}
	}

	return nil
}

func (e *Env) removeDirAlls(dirs ...string) error {
	for _, dir := range dirs {
		if err := os.RemoveAll(dir); err != nil {
			return errors.SafeWrap(fmt.Errorf("path %s: %s", dir, err), "failed to remove dir")
		}
	}

	return nil
}

func (e *Env) SetupState(depsFile string) error {
	f, err := os.Open(depsFile)
	if err != nil {
		return err
	}
	defer f.Close()

	gzr, err := gzip.NewReader(f)
	if err != nil {
		return err
	}
	defer gzr.Close()

	tr := tar.NewReader(gzr)

	for {
		header, err := tr.Next()

		switch {
		case err == io.EOF:
			return nil
		case err != nil:
			return err
		case header == nil:
			continue
		}

		target := filepath.Join(e.Config.CFDevHome, header.Name)

		switch header.Typeflag {
		case tar.TypeDir:
			if _, err := os.Stat(target); err != nil {
				if err := os.MkdirAll(target, 0755); err != nil {
					return err
				}
			}
		case tar.TypeReg:
			f, err := os.OpenFile(target, os.O_CREATE|os.O_RDWR, os.FileMode(header.Mode))
			if err != nil {
				return err
			}

			if _, err := io.Copy(f, tr); err != nil {
				return err
			}

			f.Close()
		}
	}

	return nil
}
