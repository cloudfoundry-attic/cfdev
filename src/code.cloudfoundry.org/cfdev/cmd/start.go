package cmd

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"

	"code.cloudfoundry.org/cfdev/cfanalytics"
	"code.cloudfoundry.org/cfdev/config"
	"code.cloudfoundry.org/cfdev/env"
	"code.cloudfoundry.org/cfdev/errors"
	gdn "code.cloudfoundry.org/cfdev/garden"
	"code.cloudfoundry.org/cfdev/network"
	"code.cloudfoundry.org/cfdev/process"
	"code.cloudfoundry.org/garden"
	"code.cloudfoundry.org/garden/client"
	"code.cloudfoundry.org/garden/client/connection"
	"github.com/spf13/cobra"
)

type UI interface {
	Say(message string, args ...interface{})
	Writer() io.Writer
}

type start struct {
	Exit        chan struct{}
	UI          UI
	Config      config.Config
	Registries  string
	DepsIsoPath string
	Cpus        int
	Mem         int
}

func NewStart(Exit chan struct{}, UI UI, Config config.Config) *cobra.Command {
	s := start{Exit: Exit, UI: UI, Config: Config}
	cmd := &cobra.Command{
		Use: "start",
		RunE: func(cmd *cobra.Command, args []string) error {
			return errors.SafeWrap(s.RunE(), "cf dev start")
		},
	}
	pf := cmd.PersistentFlags()
	pf.StringVarP(&s.DepsIsoPath, "file", "f", "", "path to .dev file containing bosh & cf bits")
	pf.StringVarP(&s.Registries, "registries", "r", "", "docker registries that skip ssl validation - ie. host:port,host2:port2")
	pf.IntVarP(&s.Cpus, "cpus", "c", 4, "cpus to allocate to vm")
	pf.IntVarP(&s.Mem, "memory", "m", 4096, "memory to allocate to vm in MB")

	return cmd
}

func (s *start) RunE() error {
	go func() {
		<-s.Exit
		process.SignalAndCleanup(s.Config.LinuxkitPidFile, s.Config.CFDevHome, syscall.SIGTERM)
		process.SignalAndCleanup(s.Config.VpnkitPidFile, s.Config.CFDevHome, syscall.SIGTERM)
		process.SignalAndCleanup(s.Config.HyperkitPidFile, s.Config.CFDevHome, syscall.SIGKILL)
		os.Exit(128)
	}()

	s.Config.Analytics.Event(cfanalytics.START_BEGIN, map[string]interface{}{"type": "cf"})

	if err := env.Setup(s.Config); err != nil {
		return errors.SafeWrap(err, "environment setup")
	}

	if isLinuxKitRunning(s.Config.LinuxkitPidFile) {
		s.UI.Say("CF Dev is already running...")
		s.Config.Analytics.Event(cfanalytics.START_END, map[string]interface{}{"type": "cf", "alreadyrunning": true})
		return nil
	}

	if err := cleanupStateDir(s.Config.StateDir); err != nil {
		return errors.SafeWrap(err, "cleaning state directory")
	}

	if err := s.setupNetworking(); err != nil {
		return errors.SafeWrap(err, "setting up network")
	}

	registries, err := s.parseDockerRegistriesFlag(s.Registries)
	if err != nil {
		return errors.SafeWrap(err, "Unable to parse docker registries")
	}

	vpnKit := process.VpnKit{
		Config: s.Config,
	}
	vCmd := vpnKit.Command()

	linuxkit := process.LinuxKit{
		Config:      s.Config,
		DepsIsoPath: s.DepsIsoPath,
	}

	if s.DepsIsoPath != "" {
		item := s.Config.Dependencies.Lookup("cf-deps.iso")
		item.InUse = false
	}

	if err = download(s.Config.Dependencies, s.Config.CacheDir, s.UI.Writer()); err != nil {
		return errors.SafeWrap(err, "downloading")
	}

	lCmd, err := linuxkit.Command(s.Cpus, s.Mem)
	if err != nil {
		return errors.SafeWrap(err, "Unable to find .dev file")
	}

	if !process.IsCFDevDInstalled(s.Config.CFDevDSocketPath, s.Config.CFDevDInstallationPath, s.Config.Dependencies.Lookup("cfdevd").MD5) {
		if err := process.InstallCFDevD(s.Config.CacheDir); err != nil {
			return errors.SafeWrap(err, "installing cfdevd")
		}
	}

	if err = vpnKit.SetupVPNKit(); err != nil {
		return errors.SafeWrap(err, "setting up vpnkit")
	}

	s.UI.Say("Starting VPNKit ...")
	if err := vCmd.Start(); err != nil {
		return errors.SafeWrap(err, "Failed to start VPNKit process")
	}

	err = ioutil.WriteFile(s.Config.VpnkitPidFile, []byte(strconv.Itoa(vCmd.Process.Pid)), 0777)
	if err != nil {
		return errors.SafeWrap(err, "Failed to write vpnKit pid file")
	}

	s.UI.Say("Starting the VM...")
	lCmd.Stdout, err = os.Create(filepath.Join(s.Config.CFDevHome, "linuxkit.log"))
	if err != nil {
		return errors.SafeWrap(err, "Failed to open logfile")
	}
	lCmd.Stderr = lCmd.Stdout
	if err := lCmd.Start(); err != nil {
		return errors.SafeWrap(err, "Failed to start VM process")
	}

	err = ioutil.WriteFile(s.Config.LinuxkitPidFile, []byte(strconv.Itoa(lCmd.Process.Pid)), 0777)
	if err != nil {
		return errors.SafeWrap(err, "Failed to write VM pid file")
	}

	garden := client.New(connection.New("tcp", "localhost:8888"))
	waitForGarden(garden)

	s.UI.Say("Deploying the BOSH Director...")
	if err := gdn.DeployBosh(garden); err != nil {
		return errors.SafeWrap(err, "Failed to deploy the BOSH Director")
	}

	s.UI.Say("Deploying CF...")
	if err := gdn.DeployCloudFoundry(garden, registries); err != nil {
		return errors.SafeWrap(err, "Failed to deploy the Cloud Foundry")
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

	s.Config.Analytics.Event(cfanalytics.START_END, map[string]interface{}{"type": "cf"})

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
		return errors.SafeWrap(err, "Unable to clean up .cfdev state directory")
	}

	if err := os.MkdirAll(stateDir, 0755); err != nil {
		return errors.SafeWrap(err, "Unable to create .cfdev state directory")
	}

	return nil
}

func (s *start) setupNetworking() error {
	err := network.AddLoopbackAliases(s.Config.BoshDirectorIP, s.Config.CFRouterIP)

	if err != nil {
		return errors.SafeWrap(err, "Unable to alias BOSH Director/CF Router IP")
	}

	return nil
}

func (s *start) parseDockerRegistriesFlag(flag string) ([]string, error) {
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
