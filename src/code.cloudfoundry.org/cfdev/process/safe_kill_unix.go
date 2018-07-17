// +build !windows

package process

/*
#include <libproc.h>
*/
import "C"

import (
	"io/ioutil"
	"os"
	"strconv"
	"strings"
	"syscall"
	"unsafe"
)

func (m *Manager) SafeKill(pidfile, name string) error {
	data, err := ioutil.ReadFile(pidfile)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return err
	}
	pid, err := strconv.Atoi(string(data))
	if err != nil {
		return err
	}

	path, err := executablePath(pid)
	if err != nil {
		return err
	}

	if strings.Contains(path, name) {
		syscall.Kill(pid, syscall.SIGKILL)
	}
	return os.Remove(pidfile)
}

func executablePath(pid int) (string, error) {
	var pathbuf [C.PROC_PIDPATHINFO_MAXSIZE]byte
	n := C.proc_pidpath(C.int(pid), unsafe.Pointer(&pathbuf), C.PROC_PIDPATHINFO_MAXSIZE)
	if n == 0 {
		return "", nil
	} else if n <= 0 {
		return "", syscall.ENOMEM
	}
	return string(pathbuf[:n]), nil
}
