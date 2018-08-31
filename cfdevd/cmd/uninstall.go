// +build darwin

package cmd

import (
	"net"
)

//go:generate mockgen -package mocks -destination mocks/daemonrunner.go code.cloudfoundry.org/cfdevd/cmd DaemonRunner
type DaemonRunner interface {
	RemoveDaemon(string) error
}

type UninstallCommand struct {
	DaemonRunner DaemonRunner
}

func (u *UninstallCommand) Execute(conn *net.UnixConn) error {
	err := u.DaemonRunner.RemoveDaemon("org.cloudfoundry.cfdevd")
	if err == nil {
		conn.Write([]byte{0})
	} else {
		conn.Write([]byte{1})
	}
	return err
}
