package client

import (
	"encoding/binary"
	"fmt"
	"net"
	"strings"

	"code.cloudfoundry.org/cfdev/errors"
)

type Client struct {
	name   [5]byte
	socket string
}

// clientName must be 5 characters
func New(clientName, socketPath string) *Client {
	c := &Client{socket: socketPath}
	copy(c.name[:], clientName[:5])
	return c
}

const eofReadingExitCodeMsg = "reading errorcode from cfdevd"
const connectCfdevdMsg = "connecting to cfdevd"

// sends command and returns serverName (and error)
func (c *Client) Send(command uint8) (string, error) {
	handshake := append(c.name[:], make([]byte, 44, 44)...)
	conn, err := net.Dial("unix", c.socket)
	if err != nil {
		return "", errors.SafeWrap(err, connectCfdevdMsg)
	}
	defer conn.Close()

	if err := binary.Write(conn, binary.LittleEndian, handshake); err != nil {
		return "", errors.SafeWrap(err, "sending handshake to cfdevd")
	}

	if err := binary.Read(conn, binary.LittleEndian, handshake); err != nil {
		return "", errors.SafeWrap(err, "reading handshake from cfdevd")
	}

	serverName := string(handshake[:5])
	if err := binary.Write(conn, binary.LittleEndian, []byte{command}); err != nil {
		return serverName, errors.SafeWrap(err, "sending command to cfdevd")
	}

	errorCode := make([]byte, 1, 1)
	if err := binary.Read(conn, binary.LittleEndian, errorCode); err != nil {
		return serverName, errors.SafeWrap(err, eofReadingExitCodeMsg)
	} else if errorCode[0] != 0 {
		return serverName, errors.SafeWrap(nil, fmt.Sprintf("failed to uninstall cfdevd: errorcode: %d", errorCode[0]))
	}
	return serverName, nil
}

func (c *Client) Uninstall() (string, error) {

	name, err := c.Send(1)
	if err != nil && (strings.HasPrefix(err.Error(), eofReadingExitCodeMsg) || strings.HasPrefix(err.Error(), connectCfdevdMsg)) {
		return name, nil
	}
	return name, err
}

func (c *Client) RemoveIPAlias() (string, error) {

	name, err := c.Send(2)
	if err != nil && (strings.HasPrefix(err.Error(), eofReadingExitCodeMsg) || strings.HasPrefix(err.Error(), connectCfdevdMsg)) {
		return name, nil
	}
	return name, err
}

func (c *Client) AddIPAlias() (string, error) {

	name, err := c.Send(3)
	//if err != nil && (strings.HasPrefix(err.Error(), eofReadingExitCodeMsg) || strings.HasPrefix(err.Error(), connectCfdevdMsg)) {
	//	return name, nil
	//}

	if err != nil && (strings.HasPrefix(err.Error(), eofReadingExitCodeMsg) || strings.HasPrefix(err.Error(), connectCfdevdMsg)) {

		return name, nil
	}
	return name, err
}
