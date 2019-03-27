// +build darwin

package cmd

import (
	"fmt"
	"net"
	"os"
	"os/exec"
	"strings"
)

type RemoveIPAliasCommand struct {
}

const loopback = "lo0"

func (u *RemoveIPAliasCommand) Execute(conn *net.UnixConn) error {
	err := u.RemoveLoopbackAliases(BOSH_IP, GOROUTER_IP)
	if err == nil {
		conn.Write([]byte{0})
	} else {
		conn.Write([]byte{1})
	}

	return err
}

func (u *RemoveIPAliasCommand) RemoveLoopbackAliases(addrs ...string) error {
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
