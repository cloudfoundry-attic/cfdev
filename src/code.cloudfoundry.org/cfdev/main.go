package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"syscall"

	"code.cloudfoundry.org/cfdev/process"
	"code.cloudfoundry.org/cfdev/user"
)

func main() {
	if len(os.Args) == 1 {
		fmt.Println("cfdev [start|stop]")
		os.Exit(1)
	} else if os.Args[1] == "start" {
		start()
	} else if os.Args[1] == "stop" {
		stop()
	}
}

func start() {
	devHome, err := user.CFDevHome()

	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to create .cfdev home directory: %v\n", err)
		os.Exit(1)
	}

	statePath := filepath.Join(devHome, "state")

	if err := os.MkdirAll(statePath, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "Unable to create .cfdev state directory: %v\n", err)
		os.Exit(1)
	}

	linuxkit := process.LinuxKit{
		StatePath: statePath,
		ImagePath: filepath.Join(devHome, "cfdev-efi.iso"),
	}

	cmd := linuxkit.Command()

	if err := cmd.Start(); err != nil {
		panic(err)
	}

	fmt.Printf("started linuxkit at %v\n", cmd.Process.Pid)
}

func stop() {
	devHome, _ := user.CFDevHome()
	hyperkitPid := filepath.Join(devHome, "state", "hyperkit.pid")
	pidBytes, _ := ioutil.ReadFile(hyperkitPid)
	pid, _ := strconv.ParseInt(string(pidBytes), 10, 64)

	fmt.Printf("stopping linuxkit pid %v\n", pid)
	process, _ := os.FindProcess(int(pid))
	process.Signal(syscall.SIGTERM)
}
