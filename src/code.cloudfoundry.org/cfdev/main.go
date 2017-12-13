package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"syscall"
	"time"

	gdn "code.cloudfoundry.org/cfdev/garden"
	"code.cloudfoundry.org/cfdev/network"
	"code.cloudfoundry.org/cfdev/process"
	"code.cloudfoundry.org/cfdev/resource"
	"code.cloudfoundry.org/cfdev/user"
	"code.cloudfoundry.org/garden"
	"code.cloudfoundry.org/garden/client"
	"code.cloudfoundry.org/garden/client/connection"
)

const (
	defaultDist    = "cf"
	defaultVersion = "1.2.0"
	BoshDirectorIP = "10.245.0.2"
	CFRouterIP     = "10.244.0.34"
)

func main() {
	if len(os.Args) == 1 {
		fmt.Println("cfdev [start|stop]")
		os.Exit(1)
	}

	switch os.Args[1] {
	case "start":
		start()
	case "stop":
		stop()
	case "download":
		_, _, cacheDir := setupHomeDir()
		download(cacheDir)
	default:
		fmt.Println("cfdev [start|stop]")
		os.Exit(1)
	}
}

func isSupportedVersion(flavor, version string) bool {
	return flavor == "cf" && version == "1.2.0"
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
		fmt.Fprintf(os.Stderr, "Unable to clean up .cfdev state directory: %v\n", err)
		os.Exit(1)
	}

	if err := os.MkdirAll(stateDir, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "Unable to create .cfdev state directory: %v\n", err)
		os.Exit(1)
	}
}

func download(cacheDir string) {
	fmt.Println("Downloading Resources...")
	downloader := resource.Downloader{}
	skipVerify := strings.ToLower(os.Getenv("CFDEV_SKIP_ASSET_CHECK"))

	cache := resource.Cache{
		Dir:                   cacheDir,
		DownloadFunc:          downloader.Start,
		SkipAssetVerification: skipVerify == "true",
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

func setupNetworking() {
	err := network.AddLoopbackAliases(BoshDirectorIP, CFRouterIP)

	if err != nil {
		if err == network.UnprivilegedError {
			fmt.Fprint(os.Stderr, "Please run '"+os.Args[0]+" start' "+
				"as root to setup network access to the BOSH Director/CF Router\n")
		} else {
			fmt.Fprintf(os.Stderr, "Unable to alias BOSH Director/CF Router IP: %v\n", err)
		}

		os.Exit(1)
	}

}

func start() {
	startCmd := flag.NewFlagSet("start", flag.ExitOnError)
	flavor := startCmd.String("f", defaultDist, "distribution")
	version := startCmd.String("n", defaultVersion, "version to deploy")

	startCmd.Parse(os.Args[2:])
	if !isSupportedVersion(*flavor, *version) {
		fmt.Fprintf(os.Stderr, "Distribution '%v' and version '%v' is not supported\n", *flavor, *version)
		os.Exit(1)
	}

	_, stateDir, cacheDir := setupHomeDir()
	linuxkitPidPath := filepath.Join(stateDir, "linuxkit.pid")

	if isLinuxKitRunning(linuxkitPidPath) {
		fmt.Println("CF Dev is already running...")
		return
	}

	cleanupStateDir(stateDir)
	setupNetworking()
	download(cacheDir)

	linuxkit := process.LinuxKit{
		ExecutablePath: cacheDir,
		StatePath:      stateDir,
		ImagePath:      filepath.Join(cacheDir, "cfdev-efi.iso"),
		BoshISOPath:    filepath.Join(cacheDir, "bosh-deps.iso"),
		CFISOPath:      filepath.Join(cacheDir, "cf-deps.iso"),
	}

	cmd := linuxkit.Command()
	if err := cmd.Start(); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to start LinuxKit process: %v\n", err)
		os.Exit(1)
	}

	err := ioutil.WriteFile(linuxkitPidPath, []byte(strconv.Itoa(cmd.Process.Pid)), 0777)

	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to write LinuxKit pid file: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Starting the VM...")

	garden := client.New(connection.New("tcp", "localhost:7777"))

	waitForGarden(garden)

	fmt.Println("Deploying the BOSH Director...")

	if err := gdn.DeployBosh(garden); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to deploy the BOSH Director: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Deploying CF...")

	if err := gdn.DeployCloudFoundry(garden); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to deploy the Cloud Foundry: %v\n", err)
		os.Exit(1)
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

	c := resource.Catalog{
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
			{
				URL:  "https://s3.amazonaws.com/pcfdev-development/vpnkit",
				Name: "vpnkit",
				MD5:  "de7500dea85c87d49e749c7afdc9b5fa",
				OS:   "darwin",
			},
			{
				URL:  "https://s3.amazonaws.com/pcfdev-development/hyperkit",
				Name: "hyperkit",
				MD5:  "61da21b4e82e2bf2e752d043482aa966",
				OS:   "darwin",
			},
			{
				URL:  "https://s3.amazonaws.com/pcfdev-development/linuxkit",
				Name: "linuxkit",
				MD5:  "d56da3b94c8a4146256f84f14bc41d8a",
				OS:   "darwin",
			},
			{
				URL:  "https://s3.amazonaws.com/pcfdev-development/UEFI.fd",
				Name: "UEFI.fd",
				MD5:  "2eff1c02d76fc3bde60f497ce1116b09",
			},
		},
	}

	return c.Filter(runtime.GOOS)
}
