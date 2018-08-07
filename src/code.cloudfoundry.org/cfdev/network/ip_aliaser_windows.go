package network

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
)

type HostNet struct{}

func (*HostNet) AddLoopbackAliases(addrs ...string) error {
	fmt.Println("Setting up IP aliases for the BOSH Director & CF Router (requires administrator privileges)")

	err := createSwitchIfNotExist()
	if err != nil {
		return err
	}

	for _, addr := range addrs {
		exists, err := aliasExists(addr)

		if err != nil {
			return err
		}

		if exists {
			continue
		}

		cmd := exec.Command("netsh", "interface", "ip", "add", "address", "vEthernet (cfdev)", addr, "255.255.255.255")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Stdin = os.Stdin

		if err := cmd.Run(); err != nil {
			return err
		}

		err = waitForAlias(addr)
		if err != nil {
			return err
		}
	}

	return nil
}

func (*HostNet) RemoveNetworkSwitch() error {
	exists, err := switchExists()
	if err != nil {
		return err
	}
	if !exists {
		return nil
	}

	command := exec.Command("powershell.exe", "-Command", "Remove-VMSwitch -Name cfdev -force")
	return command.Run()
}

func createSwitchIfNotExist() error {
	exists, err := switchExists()
	if err != nil {
		return err
	}
	if exists {
		return nil
	}

	command := exec.Command("powershell.exe", "-Command", "New-VMSwitch -Name cfdev -SwitchType Internal -Notes 'Switch for CF Dev Networking'")
	return command.Run()
}

func waitForAlias(addr string) error {
	done := make(chan error)
	go func() {
		for {
			if exists, err := aliasExists(addr); !exists {
				time.Sleep(1 * time.Second)
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
	case _ = <-time.After(20 * time.Second):
		return fmt.Errorf("timed out waiting for alias %s", addr)
	}
}

func aliasExists(alias string) (bool, error) {
	command := exec.Command("powershell.exe", "-Command", "ipconfig")
	output, err := command.Output()
	if err != nil {
		return false, err
	}

	return strings.Contains(string(output), alias), nil
}

func switchExists() (bool, error) {
	command := exec.Command("powershell.exe", "-Command", "Get-VMSwitch cfdev*")
	output, err := command.Output()
	if err != nil {
		return false, err
	} else if string(output) == "" {
		return false, nil
	}

	return true, nil
}
