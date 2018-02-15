package process

import (
	"syscall"
	"io/ioutil"
	"strconv"
	"fmt"
)

func Terminate(pidfile string) error {
	return signal(pidfile, syscall.SIGTERM)
}

func Kill(pidfile string) error {
	return signal(pidfile, syscall.SIGKILL)
}

func signal(pidfile string, signal syscall.Signal) error {
	pidBytes, err := ioutil.ReadFile(pidfile)
	if err != nil {
		return fmt.Errorf("failed to read pidfile %s", pidfile)
	}

	pid, err := strconv.Atoi(string(pidBytes))
	if err != nil {
		return fmt.Errorf("%s did not contain an integer", pidfile)
	}

	return syscall.Kill(pid, signal)
}


