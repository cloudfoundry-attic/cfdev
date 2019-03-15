package driver

import (
	"code.cloudfoundry.org/cfdev/config"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path/filepath"
)

func IP(cfg config.Config) (string, error) {
	var (
		macAddrPath     = filepath.Join(cfg.StateLinuxkit, "mac-addr")
		vBridgeInfoPath = filepath.Join("/var/lib/libvirt/dnsmasq/virbr0.status")
	)

	macAddr, err := ioutil.ReadFile(macAddrPath)
	if err != nil {
		return "", err
	}

	vBridgeInfo, err := ioutil.ReadFile(vBridgeInfoPath)
	if err != nil {
		return "", err
	}

	var results []struct {
		IPAddr  string `json:"ip-address"`
		MacAddr string `json:"mac-address"`
	}

	err = json.Unmarshal(vBridgeInfo, &results)
	if err != nil {
		return "", err
	}

	for _, result := range results {
		if result.MacAddr == string(macAddr) {
			return result.IPAddr, nil
		}
	}

	return "", fmt.Errorf("unable to find VM IP address from '%s'", vBridgeInfoPath)
}

