package process

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"syscall"
	"path/filepath"
	"code.cloudfoundry.org/cfdev/env"
	"encoding/json"
)

type VpnKit struct {
	HomeDir  string
	CacheDir string
	StateDir string
	BoshDirectorIP string
	CFRouterIP string
}

func (v *VpnKit) Command() *exec.Cmd {
	cmd := exec.Command(path.Join(v.CacheDir, "vpnkit"),
		"--ethernet",
		path.Join(v.HomeDir, "vpnkit_eth.sock"),
		"--port",
		path.Join(v.HomeDir, "vpnkit_port.sock"),
		"--vsock-path",
		path.Join(v.StateDir, "connect"),
		"--http",
		path.Join(v.HomeDir, "http_proxy.json"))

	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true,
	}

	return cmd
}

func (v *VpnKit) SetupVPNKit() error {
	httpProxyPath := filepath.Join(v.HomeDir, "http_proxy.json")

	proxyConfig := env.BuildProxyConfig(v.BoshDirectorIP, v.CFRouterIP)
	proxyContents, err := json.Marshal(proxyConfig)
	if err != nil {
		return fmt.Errorf("Unable to create proxy config: %v\n", err)
	}

	if _, err := os.Stat(httpProxyPath); !os.IsNotExist(err) {
		err = os.Remove(httpProxyPath)
		if err != nil {
			return fmt.Errorf("Unable to remove 'http_proxy.json' %v\n", err)
		}
	}

	httpProxyConfig := []byte(proxyContents)
	err = ioutil.WriteFile(httpProxyPath, httpProxyConfig, 0777)
	if err != nil {
		return err
	}
	return nil
}
