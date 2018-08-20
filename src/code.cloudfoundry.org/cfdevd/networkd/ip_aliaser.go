package networkd

import (
	"fmt"
	"os/exec"
	"os"
	"strings"
	"net"
)

const loopback = "lo0"

type HostNetD struct{}

func (*HostNetD) AddLoopbackAliases(addrs ...string) error {
	fmt.Println("Setting up IP aliases for the BOSH Director & CF Router (requires administrator privileges)")

	if err := createInterface(); err != nil {
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

		err = addAlias(addr)
		if err != nil {
			return err
		}
	}
	return nil
}

func (*HostNetD) RemoveLoopbackAliases(addrs ...string) error {
	for _, addr := range addrs {
		if exists, err := aliasExists(addr); err != nil {
			return err
		} else if exists {
			if err := removeAlias(addr); err != nil {
				return fmt.Errorf("removing alias %s: %s", addr, err)
			}
		}
	}
	return nil
}

func addAlias(alias string) error {
	cmd := exec.Command("sudo", "-S", "ifconfig", loopback, "add", alias+"/32")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	return cmd.Run()
}

func createInterface() error {
	return nil
}

func aliasExists(alias string) (bool, error) {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return false, fmt.Errorf("getting interface addrs: %s", err)
	}
	for _, addr := range addrs {
		if strings.Contains(addr.String(), alias) {
			return true, nil
		}
	}

	return false, nil
}

func removeAlias(alias string) error {
	cmd := exec.Command("sudo", "-S", "ifconfig", loopback, "inet", alias+"/32", "remove")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	return cmd.Run()
}

