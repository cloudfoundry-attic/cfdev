package network

import (
	"fmt"
	"net"
	"os/exec"
	"os/user"
	"strings"
)

func AddLoopbackAliases(addrs ...string) error {
	for _, addr := range addrs {
		exists, err := aliasExists(addr)
		if err != nil {
			return err
		} else if exists {
			return nil
		}

		isRoot, err := isUserRoot()
		if err != nil {
			return err
		}

		if !isRoot {
			return UnprivilegedError
		}

		cmd := exec.Command("ifconfig", "lo0", "add", addr+"/32")

		bytes, err := cmd.CombinedOutput()
		if err != nil {
			return err
		}

		if !cmd.ProcessState.Success() {
			return fmt.Errorf("unable to add alias to loopback: %s", string(bytes))
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

func isUserRoot() (bool, error) {
	u, err := user.Current()
	if err != nil {
		return false, err
	}

	if u.Uid == "0" {
		return true, nil
	}

	return false, nil
}
