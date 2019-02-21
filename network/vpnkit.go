package network

import (
	"code.cloudfoundry.org/cfdev/config"
	"code.cloudfoundry.org/cfdev/daemon"
	"code.cloudfoundry.org/cfdev/env"
	"code.cloudfoundry.org/cfdev/errors"
	"code.cloudfoundry.org/cfdev/runner"
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"
)

const VpnKitLabel = "org.cloudfoundry.cfdev.vpnkit"

type VpnKit struct {
	Config        config.Config
	DaemonRunner  DaemonRunner
	Powershell    runner.Powershell
	Label         string
	EthernetGUID  string
	PortGUID      string
	ForwarderGUID string
}

type DaemonRunner interface {
	AddDaemon(daemon.DaemonSpec) error
	RemoveDaemon(string) error
	Start(string) error
	Stop(string) error
	IsRunning(string) (bool, error)
}

func (v *VpnKit) Watch(exit chan string) {
	go func() {
		for {
			running, err := v.DaemonRunner.IsRunning(v.Label)
			if !running && err == nil {
				exit <- "vpnkit"
				return
			}
			time.Sleep(5 * time.Second)
		}
	}()
}

func (v *VpnKit) writeHttpConfig() error {
	httpProxyPath := filepath.Join(v.Config.VpnKitStateDir, "http_proxy.json")

	proxyConfig := env.BuildProxyConfig(v.Config.BoshDirectorIP, v.Config.CFRouterIP, v.Config.HostIP)
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
