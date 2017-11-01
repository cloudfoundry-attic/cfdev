package main

import (
	"fmt"
	"os/exec"

	"code.cloudfoundry.org/cli/plugin"
)

type BasicPlugin struct{}

func startLinuxKit() {

	cmd := exec.Command("sleep", "999")
	//cmd.SysProcAttr = &sysProc
	err := cmd.Start()
	if err != nil {
		fmt.Println(err)
	}
	//stdoutStderr, err := cmd.CombinedOutput()
	//fmt.Println(err)
	//fmt.Println(string(stdoutStderr))
}

func (c *BasicPlugin) Run(cliConnection plugin.CliConnection, args []string) {
	// Ensure that we called the command basic-plugin-command
	if args[1] == "start" {
		fmt.Println("Running the basic-plugin-command")

		startLinuxKit()

		fmt.Println("test")
	}
}

func (c *BasicPlugin) GetMetadata() plugin.PluginMetadata {
	return plugin.PluginMetadata{
		Name: "cfdev",
		Version: plugin.VersionType{
			Major: 0,
			Minor: 0,
			Build: 1,
		},
		MinCliVersion: plugin.VersionType{
			Major: 6,
			Minor: 7,
			Build: 0,
		},
		Commands: []plugin.Command{
			{
				Name:     "dev",
				Alias:    "dev",
				HelpText: "help text",
				// UsageDetails is optional
				// It is used to show help of usage of each command
				UsageDetails: plugin.Usage{
					Usage: "start\n   cf dev start",
				},
			},
		},
	}
}

func main() {
	// Any initialization for your plugin can be handled here
	//
	// Note: to run the plugin.Start method, we pass in a pointer to the struct
	// implementing the interface defined at "code.cloudfoundry.org/cli/plugin/plugin.go"
	//
	// Note: The plugin's main() method is invoked at install time to collect
	// metadata. The plugin will exit 0 and the Run([]string) method will not be
	// invoked.
	plugin.Start(new(BasicPlugin))
	// Plugin code should be written in the Run([]string) method,
	// ensuring the plugin environment is bootstrapped.
}
