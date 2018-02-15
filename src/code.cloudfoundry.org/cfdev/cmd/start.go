package cmd

import (
	"fmt"
	"os"
	"io/ioutil"
	"strconv"
	"path/filepath"
	"flag"
	"time"
	"syscall"
	"strings"
	"net/url"
	gdn "code.cloudfoundry.org/cfdev/garden"
	"code.cloudfoundry.org/cfdev/network"
	"code.cloudfoundry.org/cfdev/resource"
	"code.cloudfoundry.org/cfdev/user"
	"code.cloudfoundry.org/cfdev/process"
	"code.cloudfoundry.org/garden"
	"code.cloudfoundry.org/garden/client"
	"code.cloudfoundry.org/garden/client/connection"
)

const (
	BoshDirectorIP = "10.245.0.2"
	CFRouterIP     = "10.144.0.34"
)

type UI interface {
	Say(message string, args ...interface{})
}

type Start struct {
	Exit chan struct{}
	UI UI
}

func (s *Start) Run(args []string) error {
	startCmd := flag.NewFlagSet("start", flag.ExitOnError)
	registriesFlag := startCmd.String("r", "", "docker registries that skip ssl validation - ie. host:port,host2:port2")
	startCmd.Parse(args)

	homeDir, stateDir, cacheDir, err := setupHomeDir()
	if err != nil {
		return err
	}

	linuxkitPidPath := filepath.Join(stateDir, "linuxkit.pid")
	vpnkitPidPath := filepath.Join(stateDir, "vpnkit.pid")
	hyperkitPidPath := filepath.Join(stateDir, "hyperkit.pid")

	if isLinuxKitRunning(linuxkitPidPath) {
		s.UI.Say("CF Dev is already running...")
		return nil
	}

	registries, err := parseDockerRegistriesFlag(*registriesFlag)
	if err != nil {
		fmt.Errorf("Unable to parse docker registries %v\n", err)
		os.Exit(1)
	}

	vpnKit := process.VpnKit{
		HomeDir:        homeDir,
		CacheDir:       cacheDir,
		StateDir:       stateDir,
		BoshDirectorIP: BoshDirectorIP,
		CFRouterIP:     CFRouterIP,
	}
	vCmd := vpnKit.Command()

	linuxkit := process.LinuxKit{
		ExecutablePath:      cacheDir,
		StatePath:           stateDir,
		HomeDir:             homeDir,
		OSImagePath:         filepath.Join(cacheDir, "cfdev-efi.iso"),
		DependencyImagePath: filepath.Join(cacheDir, "cf-oss-deps.iso"),
	}
	lCmd := linuxkit.Command()

	go func() {
		<-s.Exit
		process.Terminate(linuxkitPidPath)
		process.Terminate(vpnkitPidPath)
		process.Kill(hyperkitPidPath)
		os.Exit(128)
	}()

	if err = cleanupStateDir(stateDir); err != nil {
		return err
	}

	if err = setupNetworking(); err != nil {
		return err
	}

	if err = s.download(cacheDir); err != nil {
		return err
	}

	if err = vpnKit.SetupVPNKit(); err != nil {
		return err
	}

	s.UI.Say("Starting VPNKit ...")
	if err := vCmd.Start(); err != nil {
		return fmt.Errorf("Failed to start VPNKit process: %v\n", err)
	}

	err = ioutil.WriteFile(vpnkitPidPath, []byte(strconv.Itoa(vCmd.Process.Pid)), 0777)
	if err != nil {
		return fmt.Errorf("Failed to write vpnKit pid file: %v\n", err)
	}

	s.UI.Say("Starting the VM...")
	if err := lCmd.Start(); err != nil {
		return fmt.Errorf("Failed to start VM process: %v\n", err)
	}

	err = ioutil.WriteFile(linuxkitPidPath, []byte(strconv.Itoa(lCmd.Process.Pid)), 0777)
	if err != nil {
		return fmt.Errorf("Failed to write VM pid file: %v\n", err)
	}

	garden := client.New(connection.New("tcp", "localhost:8888"))

	waitForGarden(garden)

	s.UI.Say("Deploying the BOSH Director...")

	if err := gdn.DeployBosh(garden); err != nil {
		return fmt.Errorf("Failed to deploy the BOSH Director: %v\n", err)
	}

	s.UI.Say("Deploying CF...")

	if err := gdn.DeployCloudFoundry(garden, registries); err != nil {
		return fmt.Errorf("Failed to deploy the Cloud Foundry: %v\n", err)
	}

	s.UI.Say(`
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
	return nil
}

func waitForGarden(client garden.Client) {
	for {
		if err := client.Ping(); err == nil {
			return
		}

		time.Sleep(time.Second)
	}
}

func setupHomeDir() (string, string, string, error) {
	homeDir, err := user.CFDevHome()

	if err != nil {
		return "", "", "", fmt.Errorf("Unable to create .cfdev home directory: %v\n", err)
	}

	stateDir := filepath.Join(homeDir, "state")

	if err := os.MkdirAll(stateDir, 0755); err != nil {
		return "", "", "", fmt.Errorf("Unable to create .cfdev state directory: %v\n", err)
	}

	cacheDir := filepath.Join(homeDir, "cache")

	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		return "", "", "", fmt.Errorf("Unable to create .cfdev cache directory: %v\n", err)
	}

	return homeDir, stateDir, cacheDir, nil
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

func cleanupStateDir(stateDir string) error {
	if err := os.RemoveAll(stateDir); err != nil {
		return fmt.Errorf("Unable to clean up .cfdev state directory: %v\n", err)
	}

	if err := os.MkdirAll(stateDir, 0755); err != nil {
		return fmt.Errorf("Unable to create .cfdev state directory: %v\n", err)
	}

	return nil
}

func setupNetworking() error {
	err := network.AddLoopbackAliases(BoshDirectorIP, CFRouterIP)

	if err != nil {
		return fmt.Errorf("Unable to alias BOSH Director/CF Router IP: %v\n", err)
	}

	return nil
}

func(s *Start) download(cacheDir string) error {
	s.UI.Say("Downloading Resources...")
	downloader := resource.Downloader{}
	skipVerify := strings.ToLower(os.Getenv("CFDEV_SKIP_ASSET_CHECK"))

	cache := resource.Cache{
		Dir:                   cacheDir,
		DownloadFunc:          downloader.Start,
		SkipAssetVerification: skipVerify == "true",
	}

	catalog, err := catalog(s.UI)
	if err != nil {
		return err
	}

	if err := cache.Sync(catalog); err != nil {
		return fmt.Errorf("Unable to sync assets: %v\n", err)
	}
	return nil
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
