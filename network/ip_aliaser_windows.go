package network

import (
	"fmt"
	"os/exec"
	"strings"
	"time"
)

const loopback = "vEthernet (cfdev)"

func (h *HostNet) RemoveLoopbackAliases(addrs ...string) error {
	exists, err := h.switchExists()
	if err != nil {
		return err
	}
	if !exists {
		return nil
	}

	_, err = h.Powershell.Output("Remove-VMSwitch -Name cfdev -force")
	return err
}

func (h *HostNet) AddLoopbackAliases(addrs ...string) error {
	fmt.Println("Setting up IP aliases for the BOSH Director & CF Router (requires administrator privileges)")

	if err := h.createSwitchIfNotExist(); err != nil {
		return err
	}

	for _, addr := range addrs {
		exists, err := h.aliasExists(addr)

		if err != nil {
			return err
		}

		if exists {
			continue
		}

		err = h.addAlias(addr)
		if err != nil {
			return err
		}
	}
	return nil
}

func (h *HostNet) addAlias(alias string) error {
	cmd := exec.Command("netsh", "interface", "ip", "add", "address", loopback, alias, "255.255.255.255")

	if err := cmd.Run(); err != nil {
		return err
	}

	return h.waitForAlias(alias)
}

func (h *HostNet) aliasExists(alias string) (bool, error) {
	output, err := h.Powershell.Output("ipconfig")
	if err != nil {
		return false, err
	}

	return strings.Contains(output, alias), nil
}

func (h *HostNet) createSwitchIfNotExist() error {
	exists, err := h.switchExists()
	if err != nil {
		return err
	}
	if exists {
		return nil
	}

	_, err = h.Powershell.Output("New-VMSwitch -Name cfdev -SwitchType Internal -Notes 'Switch for CF Dev Networking'")
	return err
}

func (h *HostNet) switchExists() (bool, error) {
	output, err := h.Powershell.Output("Get-VMSwitch cfdev*")
	if err != nil {
		return false, err
	} else if output == "" {
		return false, nil
	}

	return true, nil
}

func (h *HostNet) waitForAlias(addr string) error {
	done := make(chan error)
	go func() {
		for {
			if exists, err := h.aliasExists(addr); !exists {
				time.Sleep(3 * time.Second)
			} else if err != nil {
				done <- err
				close(done)
				return
			} else {
				close(done)
				return
			}
		}
	}()

	select {
	case err := <-done:
		return err
	case _ = <-time.After(1 * time.Minute):
		return fmt.Errorf("timed out waiting for alias %s", addr)
	}
}
