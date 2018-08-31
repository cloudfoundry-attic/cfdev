// +build darwin

package cmd

import (
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"os"
	"syscall"
)

const BOSH_IP = "10.245.0.2"
const GOROUTER_IP = "10.144.0.34"

const (
	ERROR_IN_USE    = uint8(48)
	ERROR_NOT_AVAIL = uint8(49)
	ERROR_DENIED    = uint8(71)
	ERROR_UNKNOWN   = uint8(66)
)

type BindCommand struct {
	Addr *net.TCPAddr
}

func (b *BindCommand) isIPAllowed(ip net.IP) bool {
	allowedIPs := []net.IP{
		net.ParseIP(BOSH_IP),
		net.ParseIP(GOROUTER_IP),
	}
	for _, allowed := range allowedIPs {
		if allowed.Equal(ip) {
			return true
		}
	}
	return false
}

func UnmarshalBindCommand(conn io.Reader) (*BindCommand, error) {
	ip := make([]byte, 4, 4)
	var port uint16
	var isUDP bool
	binary.Read(conn, binary.LittleEndian, ip)
	binary.Read(conn, binary.LittleEndian, &port)
	binary.Read(conn, binary.LittleEndian, &isUDP)
	if isUDP {
		return nil, fmt.Errorf("Unimplemented UDP socket requested")
	}
	return &BindCommand{
		Addr: &net.TCPAddr{
			IP:   []byte{ip[3], ip[2], ip[1], ip[0]},
			Port: int(port),
		},
	}, nil
}

func (b *BindCommand) Execute(conn *net.UnixConn) error {
	fmt.Printf("Executing tcp bind request for %s \n", b.Addr)

	msg := make([]byte, 8, 8)
	var scmsg []byte
	if b.isIPAllowed(b.Addr.IP) {
		fmt.Println("Attempting to bind address ", b.Addr)
		file, err := b.bind()
		if file != nil {
			defer file.Close()
		}
		msg, scmsg = b.response(file, err)
	} else {
		fmt.Println("Refusing to bind forbidden address ", b.Addr)
		msg[0] = ERROR_DENIED
	}
	if _, _, err := conn.WriteMsgUnix(msg, scmsg, nil); err != nil {
		return fmt.Errorf("Error writing unix msg: %s", err)
	}
	return nil
}

func (b *BindCommand) bind() (*os.File, error) {
	listener, err := net.ListenTCP("tcp", b.Addr)
	if err != nil {
		return nil, err
	}
	defer listener.Close()
	return listener.File()
}

func (b *BindCommand) response(file *os.File, err error) ([]byte, []byte) {
	msg := make([]byte, 8, 8)
	var scmsg []byte
	if err != nil {
		if opErr, ok := err.(*net.OpError); ok {
			if sysErr, ok := opErr.Err.(*os.SyscallError); ok {
				switch sysErr.Err {
				case syscall.EADDRINUSE:
					fmt.Println("Failed to Bind: address in use", b.Addr)
					msg[0] = ERROR_IN_USE
				case syscall.EADDRNOTAVAIL:
					fmt.Println("Failed to Bind: address not available", b.Addr)
					msg[0] = ERROR_NOT_AVAIL
				default:
					fmt.Println("Failed to Bind: unknown error", err)
					msg[0] = ERROR_UNKNOWN
				}
			}
		}
	}
	if file != nil {
		scmsg = syscall.UnixRights(int(file.Fd()))
	}
	return msg, scmsg
}
