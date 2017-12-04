package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"syscall"
	"time"

	gdn "code.cloudfoundry.org/cfdev/garden"
	"code.cloudfoundry.org/cfdev/process"
	"code.cloudfoundry.org/cfdev/resource"
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
	} else if os.Args[1] == "download" {
		_, _, cacheDir := setupHomeDir()
		download(cacheDir)
	}
}

func setupHomeDir() (string, string, string) {
	homeDir, err := user.CFDevHome()

	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to create .cfdev home directory: %v\n", err)
		os.Exit(1)
	}

	stateDir := filepath.Join(homeDir, "state")

	if err := os.MkdirAll(stateDir, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "Unable to create .cfdev state directory: %v\n", err)
		os.Exit(1)
	}

	cacheDir := filepath.Join(homeDir, "cache")

	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "Unable to create .cfdev cache directory: %v\n", err)
		os.Exit(1)
	}

	return homeDir, stateDir, cacheDir
}

func cleanupStateDir(stateDir string) {
	if err := os.RemoveAll(stateDir); err != nil {
		panic(err)
	}

	if err := os.MkdirAll(stateDir, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "Unable to create .cfdev state directory: %v\n", err)
		os.Exit(1)
	}
}

func download(cacheDir string) {
	fmt.Println("Downloading Resources...")

	downloader := resource.Downloader{}

	cache := resource.Cache{
		Dir:          cacheDir,
		DownloadFunc: downloader.Start,
	}

	if err := cache.Sync(catalog()); err != nil {
		fmt.Fprintf(os.Stderr, "Unable to sync assets: %v\n", err)
		os.Exit(1)
	}
}

func isLinuxKitRunning(pidFile string) bool {
	fileBytes, err := ioutil.ReadFile(pidFile)
	if err != nil {
		return false
	}

	pid, err := strconv.ParseInt(string(fileBytes), 10, 64)
	if err != nil {
		return false
	}

	if process, err := os.FindProcess(int(pid)); err == nil {
		err = process.Signal(syscall.Signal(0))
		return err == nil
	}

	return false
}

func start() {
	_, stateDir, cacheDir := setupHomeDir()
	linuxkitPidPath := filepath.Join(stateDir, "linuxkit.pid")

	if isLinuxKitRunning(linuxkitPidPath) {
		fmt.Println("CF Dev is already running...")
		return
	}

	cleanupStateDir(stateDir)
	download(cacheDir)

	linuxkit := process.LinuxKit{
		StatePath:   stateDir,
		ImagePath:   filepath.Join(cacheDir, "cfdev-efi.iso"),
		BoshISOPath: filepath.Join(cacheDir, "bosh-deps.iso"),
		CFISOPath:   filepath.Join(cacheDir, "cf-deps.iso"),
	}

	cmd := linuxkit.Command()

	if err := cmd.Start(); err != nil {
		panic(err)
	}

	err := ioutil.WriteFile(linuxkitPidPath, []byte(strconv.Itoa(cmd.Process.Pid)), 0777)

	if err != nil {
		panic(err)
	}

	fmt.Println("Starting the VM...")

	garden := client.New(connection.New("tcp", "localhost:7777"))

	waitForGarden(garden)

	fmt.Println("Deploying the BOSH Director...")

	if err := gdn.DeployBosh(garden); err != nil {
		panic(err)
	}

	fmt.Println("Deploying CF...")

	if err := gdn.DeployCloudFoundry(garden); err != nil {
		panic(err)
	}

	fmt.Println(`
  ██████╗███████╗██████╗ ███████╗██╗   ██╗
 ██╔════╝██╔════╝██╔══██╗██╔════╝██║   ██║
 ██║     █████╗  ██║  ██║█████╗  ██║   ██║
 ██║     ██╔══╝  ██║  ██║██╔══╝  ╚██╗ ██╔╝
 ╚██████╗██║     ██████╔╝███████╗ ╚████╔╝
  ╚═════╝╚═╝     ╚═════╝ ╚══════╝  ╚═══╝
             is now running!

To begin using CF Dev, please run:
    cf login -a https://api.v2.pcfdev.io --skip-ssl-validation

Admin user => Email: admin / Password: admin
Regular user => Email: user / Password: pass`)

}

func stop() {
	devHome, _ := user.CFDevHome()
	linuxkitPid := filepath.Join(devHome, "state", "linuxkit.pid")
	pidBytes, _ := ioutil.ReadFile(linuxkitPid)
	pid, _ := strconv.ParseInt(string(pidBytes), 10, 64)

	syscall.Kill(int(-pid), syscall.SIGKILL)
}

func waitForGarden(client garden.Client) {
	for {
		if err := client.Ping(); err == nil {
			return
		}

		time.Sleep(time.Second)
	}
}

func catalog() *resource.Catalog {
	override := os.Getenv("CFDEV_CATALOG")

	if override != "" {
		var c resource.Catalog
		if err := json.Unmarshal([]byte(override), &c); err != nil {
			fmt.Fprintf(os.Stderr, "Unable to parse CFDEV_CATALOG env variable: %v\n", err)
			os.Exit(1)
		}

		fmt.Println("Using CFDEV_CATALOG override")
		return &c
	}

	return &resource.Catalog{
		Items: []resource.Item{
			{
				URL:  "https://s3.amazonaws.com/pcfdev-development/cf-deps.iso",
				Name: "cf-deps.iso",
				MD5:  "81d87b3d44756518a633ba76d10be6f0",
			},
			{
				URL:  "https://s3.amazonaws.com/pcfdev-development/bosh-deps.iso",
				Name: "bosh-deps.iso",
				MD5:  "01897f5ffcee02c79d2df88ad2f4edf7",
			},
			{
				URL:  "https://s3.amazonaws.com/pcfdev-development/cfdev-efi.iso",
				Name: "cfdev-efi.iso",
				MD5:  "6a788a2a06cf0c18ac1c2ff243d223a5",
			},
		},
	}
}
