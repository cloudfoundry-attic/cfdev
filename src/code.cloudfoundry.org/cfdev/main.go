package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"syscall"
	"time"

	"code.cloudfoundry.org/cfdev/env"
	gdn "code.cloudfoundry.org/cfdev/garden"
	"code.cloudfoundry.org/cfdev/network"
	"code.cloudfoundry.org/cfdev/process"
	"code.cloudfoundry.org/cfdev/resource"
	"code.cloudfoundry.org/cfdev/shell"
	"code.cloudfoundry.org/cfdev/user"
	"code.cloudfoundry.org/garden"
	"code.cloudfoundry.org/garden/client"
	"code.cloudfoundry.org/garden/client/connection"
)

const (
	defaultDist    = "cf"
	defaultVersion = "1.2.0"
	BoshDirectorIP = "10.245.0.2"
	CFRouterIP     = "10.144.0.34"
)

func main() {
	if len(os.Args) == 1 {
		usage()
	}

	switch os.Args[1] {
	case "start":
		start()
	case "stop":
		stop()
	case "download":
		_, _, cacheDir := setupHomeDir()
		download(cacheDir)
	case "bosh":
		_, stateDir, _ := setupHomeDir()
		bosh(os.Args[2:], stateDir)
	case "catalog":
		printCatalog()
	default:
		usage()
	}
}

func usage() {
	fmt.Println("cfdev [start|stop|bosh]")
	os.Exit(1)
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
		fmt.Fprintf(os.Stderr, "Unable to alias BOSH Director/CF Router IP: %v\n", err)
		os.Exit(1)
	}
}

func setupVPNKit(homeDir string) {
	httpProxyPath := filepath.Join(homeDir, "http_proxy.json")

	proxyConfig := env.BuildProxyConfig(BoshDirectorIP, CFRouterIP)
	proxyContents, err := json.Marshal(proxyConfig)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to create proxy config: %v\n", err)
		os.Exit(1)
	}

	if _, err := os.Stat(httpProxyPath); os.IsNotExist(err) {
		err = os.Remove(httpProxyPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Unable to remove 'http_proxy.json' %v\n", err)
		}
	}

	httpProxyConfig := []byte(proxyContents)
	err = ioutil.WriteFile(httpProxyPath, httpProxyConfig, 0777)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to setup VPNKit dependencies %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("writing %s to %s", string(httpProxyConfig), httpProxyPath)
}

func start() {
	startCmd := flag.NewFlagSet("start", flag.ExitOnError)
	flavor := startCmd.String("f", defaultDist, "distribution")
	version := startCmd.String("n", defaultVersion, "version to deploy")
	registriesFlag := startCmd.String("r", "", "docker registries that skip ssl validation - ie. host:port,host2:port2")

	startCmd.Parse(os.Args[2:])
	if !isSupportedVersion(*flavor, *version) {
		fmt.Fprintf(os.Stderr, "Distribution '%v' and version '%v' is not supported\n", *flavor, *version)
		os.Exit(1)
	}

	homeDir, stateDir, cacheDir := setupHomeDir()
	linuxkitPidPath := filepath.Join(stateDir, "linuxkit.pid")
	vpnkitPidPath := filepath.Join(stateDir, "vpnkit.pid")

	if isLinuxKitRunning(linuxkitPidPath) {
		fmt.Println("CF Dev is already running...")
		return
	}

	registries, err := parseDockerRegistriesFlag(*registriesFlag)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to parse docker registries %v\n", err)
		os.Exit(1)
	}

	cleanupStateDir(stateDir)
	setupNetworking()
	download(cacheDir)
	setupVPNKit(homeDir)

	vpnKit := process.VpnKit{
		HomeDir:  homeDir,
		CacheDir: cacheDir,
		StateDir: stateDir,
	}

	fmt.Println("Starting VPNKit ...")
	vCmd := vpnKit.Command()
	if err := vCmd.Start(); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to start VPNKit process: %v\n", err)
		os.Exit(1)
	}

	err = ioutil.WriteFile(vpnkitPidPath, []byte(strconv.Itoa(vCmd.Process.Pid)), 0777)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to write vpnKit pid file: %v\n", err)
		os.Exit(1)
	}

	linuxkit := process.LinuxKit{
		ExecutablePath:      cacheDir,
		StatePath:           stateDir,
		HomeDir:             homeDir,
		OSImagePath:         filepath.Join(cacheDir, "cfdev-efi.iso"),
		DependencyImagePath: filepath.Join(cacheDir, "cf-oss-deps.iso"),
	}

	fmt.Println("Starting the VM...")
	cmd := linuxkit.Command()
	if err := cmd.Start(); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to start VM process: %v\n", err)
		os.Exit(1)
	}

	err = ioutil.WriteFile(linuxkitPidPath, []byte(strconv.Itoa(cmd.Process.Pid)), 0777)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to write VM pid file: %v\n", err)
		os.Exit(1)
	}

	garden := client.New(connection.New("tcp", "localhost:8888"))

	waitForGarden(garden)

	fmt.Println("Deploying the BOSH Director...")

	if err := gdn.DeployBosh(garden); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to deploy the BOSH Director: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Deploying CF...")

	if err := gdn.DeployCloudFoundry(garden, registries); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to deploy the Cloud Foundry: %v\n", err)
		os.Exit(1)
	}

	fmt.Print(`
  ██████╗███████╗██████╗ ███████╗██╗   ██╗
 ██╔════╝██╔════╝██╔══██╗██╔════╝██║   ██║
 ██║     █████╗  ██║  ██║█████╗  ██║   ██║
 ██║     ██╔══╝  ██║  ██║██╔══╝  ╚██╗ ██╔╝
 ╚██████╗██║     ██████╔╝███████╗ ╚████╔╝
  ╚═════╝╚═╝     ╚═════╝ ╚══════╝  ╚═══╝
             is now running!

To begin using CF Dev, please run:
    cf login -a https://api.v3.pcfdev.io --skip-ssl-validation

Admin user => Email: admin / Password: admin
Regular user => Email: user / Password: pass
`)

}

func parseDockerRegistriesFlag(flag string) ([]string, error) {
	if flag == "" {
		return nil, nil
	}

	values := strings.Split(flag, ",")

	registries := make([]string, 0, len(values))

	for _, value := range values {
		// Including the // will cause url.Parse to validate 'value' as a host:port
		u, err := url.Parse("//" + value)

		if err != nil {
			// Grab the more succinct error message
			if urlErr, ok := err.(*url.Error); ok {
				err = urlErr.Err
			}
			return nil, fmt.Errorf("'%v' - %v", value, err)
		}

		registries = append(registries, u.Host)
	}

	return registries, nil
}

func stop() {
	devHome, _ := user.CFDevHome()
	linuxkitPid := filepath.Join(devHome, "state", "linuxkit.pid")
	pidBytes, _ := ioutil.ReadFile(linuxkitPid)
	pid, _ := strconv.ParseInt(string(pidBytes), 10, 64)

	syscall.Kill(int(-pid), syscall.SIGKILL)

	vpnkitPid := filepath.Join(devHome, "state", "vpnkit.pid")
	pidBytes, _ = ioutil.ReadFile(vpnkitPid)
	pid, _ = strconv.ParseInt(string(pidBytes), 10, 64)

	syscall.Kill(int(-pid), syscall.SIGKILL)
}

func bosh(args []string, stateDir string) {
	if len(args) == 0 || args[0] != "env" {
		cmd := os.Args[0]
		fmt.Fprintf(os.Stderr, `Usage: eval $(%s bosh env)`, cmd)
		fmt.Fprintln(os.Stderr)
		os.Exit(1)
	}

	gClient := client.New(connection.New("tcp", "localhost:8888"))
	config, err := gdn.FetchBOSHConfig(gClient)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to fetch bosh configuration: %v\n", err)
		os.Exit(1)
	}

	env := shell.Environment{StateDir: stateDir}
	shellScript, err := env.Prepare(config)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to prepare bosh configuration: %v\n", err)
		os.Exit(1)
	}

	fmt.Print(shellScript)
}

func printCatalog() {
	bytes, err := json.MarshalIndent(catalog(), "", "  ")

	if err != nil {
		fmt.Fprintf(os.Stderr, "unable to marshal catalog: %v\n", err)
		os.Exit(1)
	}

	fmt.Println(string(bytes))
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
				URL:  "https://s3.amazonaws.com/pcfdev-development/stories/154480282/cf-oss-deps.iso",
				Name: "cf-oss-deps.iso",
				MD5:  "c79863e02b0ee9f984c0dd5d863d6af2",
			},
			{
				URL:  "https://s3.amazonaws.com/pcfdev-development/stories/154480282/cfdev-efi.iso",
				Name: "cfdev-efi.iso",
				MD5:  "fd1e13bb7badcacefc4e810d12a83b1d",
			},
			{
				URL:  "https://s3.amazonaws.com/pcfdev-development/stories/154480282/vpnkit",
				Name: "vpnkit",
				MD5:  "4eb4c3477e8296f4e97b5c89983d4ff3",
				OS:   "darwin",
			},
			{
				URL:  "https://s3.amazonaws.com/pcfdev-development/stories/154480282/hyperkit",
				Name: "hyperkit",
				MD5:  "61da21b4e82e2bf2e752d043482aa966",
				OS:   "darwin",
			},
			{
				URL:  "https://s3.amazonaws.com/pcfdev-development/stories/154480282/linuxkit",
				Name: "linuxkit",
				MD5:  "9ae23eec8d297f41caff3450d6a03b3c",
				OS:   "darwin",
			},
			{
				URL:  "https://s3.amazonaws.com/pcfdev-development/stories/154480282/qcow-tool",
				Name: "qcow-tool",
				MD5:  "22f3a57096ae69027c13c4933ccdd96c",
				OS:   "darwin",
			},
			{
				URL:  "https://s3.amazonaws.com/pcfdev-development/stories/154480282/UEFI.fd",
				Name: "UEFI.fd",
				MD5:  "2eff1c02d76fc3bde60f497ce1116b09",
			},
		},
	}

	return c.Filter(runtime.GOOS)
}
