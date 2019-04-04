package driver

import (
	"code.cloudfoundry.org/cfdev/config"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os/exec"
	"path/filepath"
)

func IP(cfg config.Config) (string, error) {
	var (
		ipPath          = filepath.Join(cfg.StateLinuxkit, "ip")
		macAddrPath     = filepath.Join(cfg.StateLinuxkit, "mac-addr")
		vBridgeInfoPath = "/var/lib/libvirt/dnsmasq/virbr0.status"
	)

	// The logic below is a bit of a hack.
	// Since the services will get started as root, the qemu files containing the ip address will be written as root.
	// We don't want to escalate to root every time we need the ip throughout the lifecycle of the program, so we write
	// the ip address as a normal file when we first get it. This logic is making an assumption that root privileges
	// has been retrieved as part of a prior step and has not yet timed out.
	data, err := ioutil.ReadFile(ipPath)
	if err == nil {
		return string(data), nil
	}

	macAddr, err := readAsSudo(macAddrPath)
	if err != nil {
		return "", err
	}

	vBridgeInfo, err := readAsSudo(vBridgeInfoPath)
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
			ioutil.WriteFile(ipPath, []byte(result.IPAddr), 0600)

			return result.IPAddr, nil
		}
	}

	return "", fmt.Errorf("unable to find VM IP address from '%s'", vBridgeInfoPath)
}

func readAsSudo(path string) ([]byte, error) {
	return exec.Command("sudo", "-S", "cat", path).Output()
}
