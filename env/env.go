package env

import (
	"code.cloudfoundry.org/cfdev/resource"
	"fmt"
	"io"
	"io/ioutil"
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

	if err := os.RemoveAll(config.StateDir); err != nil {
		return errors.SafeWrap(fmt.Errorf("path %s: %s", config.StateDir, err), "failed to clean up state dir")
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
	f, err := os.Open(filepath.Join(config.CacheDir, "cfdev-deps.tgz"))
	if err != nil {
		return nil
	}
	defer f.Close()

	err = resource.Untar(config.StateLinuxkit, f, resource.TarOpts{Include: "disk.qcow2"})
	if err != nil {
		return err
	}

	err = moveFile(filepath.Join(config.CacheDir, "state.json"), filepath.Join(config.StateBosh, "state.json"))
	if err != nil {
		return err
	}

	err = moveFile(filepath.Join(config.CacheDir, "creds.yml"), filepath.Join(config.StateBosh, "creds.yml"))
	if err != nil {
		return err
	}

	err = moveFile(filepath.Join(config.CacheDir, "secret"), filepath.Join(config.StateBosh, "secret"))
	if err != nil {
		return err
	}

	err = moveFile(filepath.Join(config.CacheDir, "jumpbox.key"), filepath.Join(config.StateBosh, "jumpbox.key"))
	if err != nil {
		return err
	}

	err = moveFile(filepath.Join(config.CacheDir, "ca.crt"), filepath.Join(config.StateBosh, "ca.crt"))
	if err != nil {
		return err
	}

	err = copyDir(filepath.Join(config.CacheDir, "services"), filepath.Join(config.CFDevHome, "services"))
	if err != nil {
		return err
	}

	return nil
}

func moveFile(srcDir, targetDir string) error {
	_, err := os.Stat(srcDir)
	if os.IsNotExist(err) {
		return nil
	}

	src, err := os.Open(srcDir)
	if err != nil {
		return err
	}
	defer src.Close()

	dst, err := os.Create(targetDir)
	if err != nil {
		return err
	}
	defer dst.Close()

	_, err = io.Copy(dst, src)
	if err != nil {
		return err
	}

	return err
}

func copyDir(src string, dst string) (err error) {
	src = filepath.Clean(src)
	dst = filepath.Clean(dst)

	si, err := os.Stat(src)
	if err != nil {
		return err
	}
	if !si.IsDir() {
		return fmt.Errorf("source is not a directory")
	}

	_, err = os.Stat(dst)
	if err != nil && !os.IsNotExist(err) {
		return
	}
	if err == nil {
		return fmt.Errorf("destination already exists")
	}

	err = os.MkdirAll(dst, si.Mode())
	if err != nil {
		return
	}

	entries, err := ioutil.ReadDir(src)
	if err != nil {
		return
	}

	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())

		if entry.IsDir() {
			err = copyDir(srcPath, dstPath)
			if err != nil {
				return
			}
		} else {
			// Skip symlinks.
			if entry.Mode()&os.ModeSymlink != 0 {
				continue
			}

			err = moveFile(srcPath, dstPath)
			if err != nil {
				return
			}
		}
	}

	return
}
