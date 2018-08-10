// +build darwin

package cmd

import (
	"net"
)

//go:generate mockgen -package mocks -destination mocks/launchd.go code.cloudfoundry.org/cfdevd/cmd Launchd
type Launchd interface {
	RemoveDaemon(string) error
}

type UninstallCommand struct {
	Launchd Launchd
}

func (u *UninstallCommand) Execute(conn *net.UnixConn) error {
	err := u.Launchd.RemoveDaemon("org.cloudfoundry.cfdevd")
	if err == nil {
		conn.Write([]byte{0})
	} else {
		conn.Write([]byte{1})
	}
	return err
}
