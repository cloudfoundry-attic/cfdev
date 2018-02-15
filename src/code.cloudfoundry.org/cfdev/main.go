package main

import (
	"fmt"
	"os"
	"code.cloudfoundry.org/cfdev/cmd"
	"code.cloudfoundry.org/cli/plugin"
	"os/signal"
	"syscall"
)

func main() {
	exitChan := make(chan struct{})
	sigChan := make(chan os.Signal, 1)
	signal.Notify(make(chan os.Signal), syscall.SIGHUP)
	signal.Notify(sigChan, syscall.SIGINT)
	signal.Notify(sigChan, syscall.SIGTERM)

	go func() {
		<-sigChan
		close(exitChan)
	}()

	cfdev := &Plugin{exitChan}

	plugin.Start(cfdev)
}

type Command interface {
	Run(args []string) error
}

type Plugin struct {
	Exit chan struct{}
}

func (p *Plugin) Run(connection plugin.CliConnection, args []string) {
	if args[0] == "CLI-MESSAGE-UNINSTALL" {
		return
	}
	p.execute(args[1:])
}

func (p *Plugin) GetMetadata() plugin.PluginMetadata {
	return plugin.PluginMetadata{
		Name: "cfdev",
		Version: plugin.VersionType{
			Major: 0,
			Minor: 0,
			Build: 1,
		},
		Commands: []plugin.Command{
			{
				Name:     "dev",
				HelpText: "Start and stop a single vm CF deployment running on your workstation",
				UsageDetails: plugin.Usage{
					Usage: "cfdev [start|stop|bosh]",
				},
			},
		},
	}
}

func usage() {
	fmt.Println("cfdev [start|stop|bosh]")
	os.Exit(1)
}

func (p *Plugin) execute(args []string) {
	if len(args) == 0 {
		usage()
	}

	var command Command
	switch args[0] {
	case "start":
		command = &cmd.Start{
			Exit: p.Exit,
		}
	case "stop":
		command = &cmd.Stop{}
	case "download":
		command = &cmd.Download{
			Exit: p.Exit,
		}
	case "bosh":
		command = &cmd.Bosh{
			Exit: p.Exit,
		}
	case "catalog":
		command = &cmd.Catalog{}
	default:
		usage()
	}

	command.Run(args[1:])
}
