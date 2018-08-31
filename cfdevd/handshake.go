// +build darwin

package main

import (
	"encoding/binary"
	"fmt"
	"io"
	"net"
)

type Handshake struct {
	Magic   string
	Version uint32
	Commit  string
}

func unmarshalHandshake(data []byte) Handshake {
	return Handshake{
		Magic:   string(data[0:5]),
		Version: binary.LittleEndian.Uint32(data[5:9]),
		Commit:  string(data[9:40]),
	}
}

func marshalHandshake(hsk Handshake) []byte {
	data := []byte(hsk.Magic)
	data = append(data, make([]byte, 4, 4)...)
	binary.LittleEndian.PutUint32(data[5:], hsk.Version)
	return append(data, []byte(hsk.Commit)...)
}

func doHandshake(conn net.Conn) error {
	init := make([]byte, 49, 49)
	_, err := io.ReadFull(conn, init)
	if err != nil {
		return err
	}
	handshake := unmarshalHandshake(init)
	fmt.Printf("connection received from client: Name: %s, Version: %d, Commit: %s \n",
		handshake.Magic,
		handshake.Version,
		handshake.Commit,
	)

	conn.Write(marshalHandshake(
		Handshake{
			Magic:   "CFD3V",
			Version: 22,
			Commit:  "0123456789012345678901234567890123456789",
		},
	))
	return nil
}
