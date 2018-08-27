package cmd

import (
	"net"

	"code.cloudfoundry.org/cfdev/cfdevd/networkd"
)

type AddIPAliasCommand struct {
}

func (u *AddIPAliasCommand) Execute(conn *net.UnixConn) error {

	hostNet := &networkd.HostNetD{}

	err := hostNet.AddLoopbackAliases(BOSH_IP, GOROUTER_IP)
	if err == nil {
		conn.Write([]byte{0})
	} else {
		conn.Write([]byte{1})
	}

	return nil
}
