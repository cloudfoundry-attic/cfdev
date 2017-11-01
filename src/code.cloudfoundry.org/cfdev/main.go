package main

import (
	"fmt"
	"os"
	"path/filepath"

	"code.cloudfoundry.org/cfdev/process"
	"code.cloudfoundry.org/cfdev/user"
)

func main() {
	devHome, err := user.CFDevHome()

	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to create .cfdev home directory: %v\n", err)
		os.Exit(1)
	}

	linuxkit := process.LinuxKit{
		StatePath: filepath.Join(devHome, "state"),
		ImagePath: filepath.Join(devHome, "cfdev-efi.iso"),
	}

	cmd := linuxkit.Command()

	if err := cmd.Start(); err != nil {
		panic(err)
	}

	fmt.Printf("started linuxkit at %v\n", cmd.Process.Pid)
}
