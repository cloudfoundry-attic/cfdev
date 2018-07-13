package network

import (
	"net"
	"strings"
	"fmt"
	"os/exec"
	"os"
)

type HostNet struct {}

func (*HostNet) AddLoopbackAliases(addrs ...string) error {
	prompted := false

	for _, addr := range addrs {
		exists, err := aliasExists(addr)

		if err != nil {
			return err
		}

		if exists {
			continue
		}

		if !prompted {
			fmt.Println("Setting up IP aliases for the BOSH Director & CF Router (requires administrator privileges)")
			prompted = true
		}

		err = createSwitchIfNotExist()
		if err != nil {
			return err
		}

		cmd := exec.Command("netsh", "interface", "ip", "add", "address", "vEthernet (cfdev)", addr, "255.255.255.255")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Stdin = os.Stdin

		if err := cmd.Run(); err != nil {
			return err
		}
	}

	return nil
}

func (*HostNet) RemoveNetworkSwitch() error  {
	command := exec.Command("powershell.exe", "-Command", "Remove-VMSwitch -Name cfdev -force")
	return command.Run()
}

func createSwitchIfNotExist() error {
	command := exec.Command("powershell.exe", "-Command", "Get-VMSwitch cfdev*")
	output, err := command.Output()
	if err != nil {
		return err
	}

	if string(output) != "" {
		return nil
	}

	command = exec.Command("powershell.exe", "-Command", "New-VMSwitch -Name cfdev -SwitchType Internal -Notes 'Switch for CF Dev Networking'")
	return command.Run()
}

func aliasExists(alias string) (bool, error) {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return false, err
	}

	for _, addr := range addrs {
		if strings.Contains(addr.String(), alias) {
			return true, nil
		}
	}

	return false, nil
}