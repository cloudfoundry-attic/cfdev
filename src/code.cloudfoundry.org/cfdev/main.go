package main

import (
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"syscall"
	"time"

	gdn "code.cloudfoundry.org/cfdev/garden"
	"code.cloudfoundry.org/cfdev/process"
	"code.cloudfoundry.org/cfdev/user"
	"code.cloudfoundry.org/garden"
	"code.cloudfoundry.org/garden/client"
	"code.cloudfoundry.org/garden/client/connection"
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
	if err := os.RemoveAll(statePath); err != nil {
		panic(err)
	}

	if err := os.MkdirAll(statePath, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "Unable to create .cfdev state directory: %v\n", err)
		os.Exit(1)
	}

	linuxkit := process.LinuxKit{
		StatePath:   statePath,
		ImagePath:   filepath.Join(devHome, "cfdev-efi.iso"),
		BoshISOPath: filepath.Join(devHome, "bosh-deps.iso"),
	}

	cmd := linuxkit.Command()

	if err := cmd.Start(); err != nil {
		panic(err)
	}

	linuxkitPid := filepath.Join(statePath, "linuxkit.pid")

	err = ioutil.WriteFile(linuxkitPid, []byte(strconv.Itoa(cmd.Process.Pid)), 0777)

	if err != nil {
		panic(err)
	}

	fmt.Println("Starting the VM...")

	garden := client.New(connection.New("tcp", "127.0.0.1:7777"))

	waitForGarden(garden)

	fmt.Println("Deploying the BOSH Director...")

	if err := gdn.DeployBosh(garden); err != nil {
		panic(err)
	}

	waitForBOSH()
}

func stop() {
	devHome, _ := user.CFDevHome()
	linuxkitPid := filepath.Join(devHome, "state", "linuxkit.pid")
	pidBytes, _ := ioutil.ReadFile(linuxkitPid)
	pid, _ := strconv.ParseInt(string(pidBytes), 10, 64)

	syscall.Kill(int(-pid), syscall.SIGKILL)
}

func waitUntilListeningAt(addr string) {
	for {
		fmt.Printf("Waiting to you hear back from %v\n", addr)
		_, err := net.Dial("tcp", addr)

		if err == nil {
			return
		}

		time.Sleep(time.Second)
	}
}

func waitForBOSH() {
	waitUntilListeningAt("localhost:25555")
}

func waitForGarden(client garden.Client) {
	for {
		if err := client.Ping(); err == nil {
			return
		}

		time.Sleep(time.Second)
	}
}
