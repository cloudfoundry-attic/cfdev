package start

import (
	"github.com/spf13/cobra"
	"code.cloudfoundry.org/cfdev/errors"
	"os"
	"path/filepath"
	"code.cloudfoundry.org/cfdev/cfanalytics"
	"code.cloudfoundry.org/cfdev/env"
)

func (s *Start) Cmd() *cobra.Command {
	args := Args{}
	cmd := &cobra.Command{
		Use: "start",
		RunE: func(_ *cobra.Command, _ []string) error {
			if err := s.Execute(args); err != nil {
				return errors.SafeWrap(err, "cf dev start")
			}
			return nil
		},
	}

	pf := cmd.PersistentFlags()
	pf.StringVarP(&args.DepsIsoPath, "file", "f", "", "path to .dev file containing bosh & cf bits")
	pf.StringVarP(&args.Registries, "registries", "r", "", "docker registries that skip ssl validation - ie. host:port,host2:port2")
	pf.IntVarP(&args.Cpus, "cpus", "c", 4, "cpus to allocate to vm")
	pf.IntVarP(&args.Mem, "memory", "m", 4096, "memory to allocate to vm in MB")

	return cmd
}

func (s *Start) Execute(args Args) error {
	go func() {
		select {
		case <-s.Exit:
			// no-op
		case name := <-s.LocalExit:
			s.UI.Say("ERROR: %s has stopped", name)
		}
		//TODO: HYPER-V STOP
		//s.LinuxKit.Stop()
		s.VpnKit.Stop()
		os.Exit(128)
	}()

	depsIsoName := "cf"
	depsIsoPath := filepath.Join(s.Config.CacheDir, "cf-deps.iso")
	if args.DepsIsoPath != "" {
		depsIsoName = filepath.Base(args.DepsIsoPath)
		var err error
		depsIsoPath, err = filepath.Abs(args.DepsIsoPath)
		if err != nil {
			return errors.SafeWrap(err, "determining absolute path to deps iso")
		}
	}
	s.AnalyticsToggle.SetProp("type", depsIsoName)
	s.Analytics.Event(cfanalytics.START_BEGIN)
	_ = depsIsoPath
	//if running, err := s.LinuxKit.IsRunning(); err != nil {
	//	return errors.SafeWrap(err, "is linuxkit running")
	//} else if running {
	//	s.UI.Say("CF Dev is already running...")
	//	s.Analytics.Event(cfanalytics.START_END, map[string]interface{}{"alreadyrunning": true})
	//	return nil
	//}

	if err := env.Setup(s.Config); err != nil {
		return errors.SafeWrap(err, "environment setup")
	}

	//TODO CLEANUP STATE DIR?????

	if err := s.HostNet.AddLoopbackAliases(s.Config.BoshDirectorIP, s.Config.CFRouterIP); err != nil {
		return errors.SafeWrap(err, "adding aliases")
	}

	s.UI.Say("Downloading Resources...")
	if err := s.Cache.Sync(s.Config.Dependencies); err != nil {
		return errors.SafeWrap(err, "Unable to sync assets")
	}

	s.UI.Say("Creating VM...")
	if err := s.HyperV.CreateVM(); err != nil {
		return errors.SafeWrap(err, "Unable to create VM")
	}

	s.UI.Say("Starting VPNKit...")
	if err := s.VpnKit.Start(); err != nil {
		return errors.SafeWrap(err, "starting vpnkit")
	}
	//s.VpnKit.Watch(s.LocalExit) what is this?

	s.UI.Say("Starting VM...")
	if err := s.HyperV.Start("cfdev"); err != nil {
		return errors.SafeWrap(err, "starting vpnkit")
	}

	return nil
}