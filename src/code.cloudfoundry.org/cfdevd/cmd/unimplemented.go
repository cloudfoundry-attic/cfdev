package cmd

import (
	"net"
	"fmt"
)

type UnimplementedCommand struct {
	Instruction uint8
}

func(u *UnimplementedCommand) Execute(conn *net.UnixConn) error {
	message := []byte{uint8(33)}
	fmt.Println("Unimplemented command: ", u.Instruction)
	if _, err := conn.Write(message); err != nil {
		return fmt.Errorf("unimplememted instruction: failed to write error code to connection: %s", err)
	}
	return nil
}