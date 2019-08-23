package hyperv

import (
	"fmt"
	"os/exec"
	"strings"
	"time"
)

func (d *HyperV) AddLoopbackAliases(switchName string, addrs ...string) error {
	err := d.createSwitchIfNotExist(switchName)
	if err != nil {
		return err
	}

	for _, addr := range addrs {
		exists, err := d.aliasExists(addr)
		if err != nil {
			return err
		}

		if exists {
			continue
		}

		err = d.addAlias(switchName, addr)
		if err != nil {
			return err
		}
	}

	return nil
}

func (d *HyperV) RemoveLoopbackAliases(switchName string, addrs ...string) error {
	exists, err := d.switchExists(switchName)
	if err != nil {
		return err
	}
	if !exists {
		return nil
	}

	_, err = d.Powershell.Output(fmt.Sprintf("Hyper-V\\Remove-VMSwitch -Name %s -force", switchName))
	return err
}

func (d *HyperV) loopback(switchName string) string {
	return fmt.Sprintf("vEthernet (%s)", switchName)
}


func (d *HyperV) addAlias(switchName, alias string) error {
	cmd := exec.Command("netsh", "interface", "ip", "add", "address", d.loopback(switchName), alias, "255.255.255.255")

	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to add network alias: %s, %s, %s, %s", d.loopback(switchName), alias, err, output)
	}

	return d.waitForAlias(alias)
}

func (d *HyperV) aliasExists(alias string) (bool, error) {
	output, err := d.Powershell.Output("ipconfig")
	if err != nil {
		return false, err
	}

	return strings.Contains(output, alias), nil
}

func (d *HyperV) createSwitchIfNotExist(switchName string) error {
	exists, err := d.switchExists(switchName)
	if err != nil {
		return err
	}
	if exists {
		return nil
	}

	_, err = d.Powershell.Output(fmt.Sprintf("Hyper-V\\New-VMSwitch -Name %s -SwitchType Internal -Notes 'Switch for CF Dev Networking'", switchName))
	return err
}

func (d *HyperV) switchExists(switchName string) (bool, error) {
	output, err := d.Powershell.Output(fmt.Sprintf("Hyper-V\\Get-VMSwitch %s*", switchName))
	if err != nil {
		return false, err
	} else if output == "" {
		return false, nil
	}

	return true, nil
}

func (d *HyperV) waitForAlias(addr string) error {
	var (
		ticker  = time.NewTicker(3*time.Second)
		timeout = time.After(time.Minute)
		err     error
	)

	for {
		select {
		case <-ticker.C:
			var exists bool
			exists, err = d.aliasExists(addr)
			if exists {
				return nil
			}
		case <-timeout:
			return fmt.Errorf("timed out waiting for alias: %s", err)
		}
	}
}

