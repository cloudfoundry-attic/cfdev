package cmd

import (
	"encoding/binary"
	"io"
	"net"

	"code.cloudfoundry.org/cfdevd/launchd"
	"os"
)

type Command interface {
	Execute(*net.UnixConn) error
}

const UninstallType = uint8(1)
const BindType = uint8(6)

func UnmarshalCommand(conn io.Reader) (Command, error) {
	var instr uint8
	binary.Read(conn, binary.LittleEndian, &instr)

	switch instr {
	case BindType:
		return UnmarshalBindCommand(conn)
	case UninstallType:
		return &UninstallCommand{
			Launchd: launchd.New(""),
		}, nil
	default:
		return &UnimplementedCommand{
			Instruction: instr,
			Logger: os.Stdout,
		}, nil
	}
}
