package cmd

import (
	"net"
	"fmt"
	"io"
)

type UnimplementedCommand struct {
	Instruction uint8
	Logger io.Writer
}

func(u *UnimplementedCommand) Execute(conn *net.UnixConn) error {
	message := []byte{uint8(33)}
	fmt.Fprintf(u.Logger,"Unimplemented command: %v\n", u.Instruction)
	if _, err := conn.Write(message); err != nil {
		return fmt.Errorf("unimplememted instruction: failed to write error code to connection: %s", err)
	}
	return nil
}