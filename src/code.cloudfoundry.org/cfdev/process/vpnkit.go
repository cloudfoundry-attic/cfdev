package process

import (
	"code.cloudfoundry.org/cfdev/config"
	"time"
	"path/filepath"
	"code.cloudfoundry.org/cfdev/env"
	"encoding/json"
	"code.cloudfoundry.org/cfdev/errors"
	"os"
	"io/ioutil"
	"code.cloudfoundry.org/cfdev/daemon"
)

const VpnKitLabel = "org.cloudfoundry.cfdev.vpnkit"

type VpnKit struct {
	Config  config.Config
	DaemonRunner DaemonRunner
}

type DaemonRunner interface {
	AddDaemon(daemon.DaemonSpec) error
	RemoveDaemon(string) error
	Start(string) error
	Stop(string) error
	IsRunning(string) (bool, error)
}

func (v *VpnKit) Stop() error {
	return v.DaemonRunner.Stop(VpnKitLabel)
}

func (v *VpnKit) Watch(exit chan string) {
	go func() {
		for {
			running, err := v.DaemonRunner.IsRunning(VpnKitLabel)
			if !running && err == nil {
				exit <- "vpnkit"
				return
			}
			time.Sleep(5 * time.Second)
		}
	}()
}

func (v *VpnKit) writeHttpConfig() error{
	httpProxyPath := filepath.Join(v.Config.VpnKitStateDir, "http_proxy.json")

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

