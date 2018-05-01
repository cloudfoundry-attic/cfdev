package process

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"

	"code.cloudfoundry.org/cfdev/errors"
)

func SignalAndCleanup(pidfile, match string, signal syscall.Signal) error {
	pidBytes, err := ioutil.ReadFile(pidfile)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("failed to read pidfile %s", pidfile)
	}

	pid, err := strconv.Atoi(string(pidBytes))
	if err != nil {
		return fmt.Errorf("%s did not contain an integer", pidfile)
	}

	if running, err := isRunning(pid); err != nil {
		return fmt.Errorf("failed to send signal to %s", filepath.Base(pidfile))
	} else if !running {
		return os.Remove(pidfile)
	}

	// get process form pid
	cmd := exec.Command("ps", "-p", string(pidBytes), "-o", "command=")
	output, err := cmd.Output()
	if err == nil && !strings.Contains(string(output), match) {
		// Process pid is running, but is not us
		return os.Remove(pidfile)
	}

	if err := syscall.Kill(pid, signal); err != nil {
		return fmt.Errorf("failed to send signal to %s", filepath.Base(pidfile))
	}

	if err := waitForPidTermination(pid); err != nil {

		return err
	}
	return os.Remove(pidfile)
}

func isRunning(pid int) (bool, error) {
	err := syscall.Kill(pid, syscall.Signal(0))
	if err != nil {
		if e, ok := err.(syscall.Errno); ok && e == syscall.ESRCH {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func waitForPidTermination(pid int) error {
	timeout := time.After(time.Minute)
	tick := time.NewTicker(time.Millisecond)
	defer tick.Stop()
	for {
		select {
		case <-timeout:
			return errors.SafeWrap(nil, "timed out waiting for process to exit")
		case <-tick.C:
			err := syscall.Kill(pid, syscall.Signal(0))
			if err != nil {
				if e, ok := err.(syscall.Errno); ok && e == syscall.ESRCH {
					return nil
				}
				return err
			}
		}
	}
	return nil
}
