package process

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"syscall"

	"code.cloudfoundry.org/cfdev/config"
	"code.cloudfoundry.org/cfdev/env"
	"code.cloudfoundry.org/cfdev/errors"
)

type VpnKit struct {
	Config config.Config
}

func (v *VpnKit) Command() *exec.Cmd {
	cmd := exec.Command(path.Join(v.Config.CacheDir, "vpnkit"),
		"--ethernet",
		path.Join(v.Config.CFDevHome, "vpnkit_eth.sock"),
		"--port",
		path.Join(v.Config.CFDevHome, "vpnkit_port.sock"),
		"--vsock-path",
		path.Join(v.Config.StateDir, "connect"),
		"--http",
		path.Join(v.Config.CFDevHome, "http_proxy.json"))

	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true,
	}

	return cmd
}

func (v *VpnKit) SetupVPNKit() error {
	httpProxyPath := filepath.Join(v.Config.CFDevHome, "http_proxy.json")

	proxyConfig := env.BuildProxyConfig(v.Config.BoshDirectorIP, v.Config.CFRouterIP)
	proxyContents, err := json.Marshal(proxyConfig)
	if err != nil {
		return errors.SafeWrap(err, "Unable to create proxy config")
	}

	if _, err := os.Stat(httpProxyPath); !os.IsNotExist(err) {
		err = os.Remove(httpProxyPath)
		if err != nil {
			return errors.SafeWrap(err, "Unable to remove 'http_proxy.json'")
		}
	}

	httpProxyConfig := []byte(proxyContents)
	err = ioutil.WriteFile(httpProxyPath, httpProxyConfig, 0777)
	if err != nil {
		return err
	}
	return nil
}
