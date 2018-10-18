package network

import (
	"fmt"
	"os/exec"
	"strings"
	"time"
)

func (h *HostNet) loopback() string {
	return fmt.Sprintf("vEthernet (%s)", h.VMSwitchName)
}
func (h *HostNet) RemoveLoopbackAliases(addrs ...string) error {
	exists, err := h.switchExists()
	if err != nil {
		return err
	}
	if !exists {
		return nil
	}

	_, err = h.Powershell.Output(fmt.Sprintf("Remove-VMSwitch -Name %s -force", h.VMSwitchName))
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
	cmd := exec.Command("netsh", "interface", "ip", "add", "address", h.loopback(), alias, "255.255.255.255")

	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to add network alias: %s, %s, %s, %s", h.loopback(), alias, err, output)
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

	_, err = h.Powershell.Output(fmt.Sprintf("New-VMSwitch -Name %s -SwitchType Internal -Notes 'Switch for CF Dev Networking'", h.VMSwitchName))
	return err
}

func (h *HostNet) switchExists() (bool, error) {
	output, err := h.Powershell.Output(fmt.Sprintf("Get-VMSwitch %s*", h.VMSwitchName))
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
