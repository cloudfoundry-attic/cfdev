package cmd

import (
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"sync"
	"syscall"

	"code.cloudfoundry.org/cfdev/config"
	"code.cloudfoundry.org/cfdev/process"
)

type Stop struct {
	Config config.Config
}

func (s *Stop) Run(args []string) error {
	var reterr error
	var all sync.WaitGroup
	all.Add(4)

	go func() {
		defer all.Done()
		if err := process.SignalAndCleanup(s.Config.LinuxkitPidFile, s.Config.CFDevHome, syscall.SIGTERM); err != nil {
			reterr = fmt.Errorf("failed to terminate linuxkit: %s", err)
		}
	}()
	go func() {
		defer all.Done()
		if err := process.SignalAndCleanup(s.Config.VpnkitPidFile, s.Config.CFDevHome, syscall.SIGTERM); err != nil {
			reterr = fmt.Errorf("failed to terminate vpnkit: %s", err)
		}
	}()
	go func() {
		defer all.Done()
		if err := process.SignalAndCleanup(s.Config.HyperkitPidFile, s.Config.CFDevHome, syscall.SIGKILL); err != nil {
			reterr = fmt.Errorf("failed to terminate hyperkit: %s", err)
		}
	}()
	go func() {
		defer all.Done()
		command := []byte{uint8(1)}
		handshake := append([]byte("CFD3V"), make([]byte, 44, 44)...)
		conn, err := net.Dial("unix", s.Config.CFDevDSocketPath)
		if err != nil {
			// cfdevd is not running-- do nothing
			return
		}
		if err := binary.Write(conn, binary.LittleEndian, handshake); err != nil {
			reterr = err
			return
		}
		if err := binary.Read(conn, binary.LittleEndian, handshake); err != nil {
			reterr = err
			return
		}
		if err := binary.Write(conn, binary.LittleEndian, command); err != nil {
			reterr = err
			return
		}
		errorCode := make([]byte, 1, 1)
		if err := binary.Read(conn, binary.LittleEndian, errorCode); err != nil {
			if err != io.EOF {
				reterr = err
				return
			}
		} else if errorCode[0] != 0 {
			reterr = fmt.Errorf("failed to uninstall cfdevd: errorcode: %d", errorCode[0])
		}
	}()

	all.Wait()

	return reterr
}
