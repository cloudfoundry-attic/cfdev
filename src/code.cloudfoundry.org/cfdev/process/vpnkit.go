package process

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"

	"code.cloudfoundry.org/cfdev/config"
	"code.cloudfoundry.org/cfdev/env"
	"code.cloudfoundry.org/cfdev/errors"
	launchd "code.cloudfoundry.org/cfdevd/launchd/models"
)

type VpnKit struct {
	Config config.Config
}

const VpnKitLabel = "org.cloudfoundry.cfdev.vpnkit"

func (v *VpnKit) DaemonSpec() launchd.DaemonSpec {
	return launchd.DaemonSpec{
		Label:       VpnKitLabel,
		Program:     path.Join(v.Config.CacheDir, "vpnkit"),
		SessionType: "Background",
		ProgramArguments: []string{
			path.Join(v.Config.CacheDir, "vpnkit"),
			"--ethernet",
			path.Join(v.Config.VpnkitStateDir, "vpnkit_eth.sock"),
			"--port",
			path.Join(v.Config.VpnkitStateDir, "vpnkit_port.sock"),
			"--vsock-path",
			path.Join(v.Config.StateDir, "connect"),
			"--http",
			path.Join(v.Config.VpnkitStateDir, "http_proxy.json"),
		},
		RunAtLoad:  false,
		StdoutPath: path.Join(v.Config.CFDevHome, "vpnkit.stdout.log"),
		StderrPath: path.Join(v.Config.CFDevHome, "vpnkit.stderr.log"),
	}
}

func (v *VpnKit) SetupVPNKit() error {
	httpProxyPath := filepath.Join(v.Config.VpnkitStateDir, "http_proxy.json")

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
