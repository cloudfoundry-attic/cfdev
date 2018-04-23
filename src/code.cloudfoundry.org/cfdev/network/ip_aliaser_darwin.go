package network

import (
	"fmt"
	"net"
	"os"
	"os/exec"
	"strings"
)

func AddLoopbackAliases(addrs ...string) error {
	prompted := false
	for _, addr := range addrs {
		exists, err := aliasExists(addr)

		if err != nil {
			return err
		} else if exists {
			continue
		}

		if !prompted {
			fmt.Println("Setting up IP aliases for the BOSH Director & CF Router (requires root privileges)")
			prompted = true
		}

		cmd := exec.Command("sudo", "-S", "ifconfig", "lo0", "add", addr+"/32")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Stdin = os.Stdin

		if err := cmd.Run(); err != nil {
			return err
		}
	}

	return nil
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
