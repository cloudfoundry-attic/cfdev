package cmd

import (
	"strings"

	"code.cloudfoundry.org/cfdev/config"
	"code.cloudfoundry.org/cfdev/process"
	cfdevdClient "code.cloudfoundry.org/cfdevd/client"
	"github.com/spf13/cobra"
)

func NewRoot(Exit chan struct{}, UI UI, Config config.Config, Launchd Launchd) *cobra.Command {
	root := &cobra.Command{Use: "cf", SilenceUsage: true, SilenceErrors: true}
	root.PersistentFlags().Bool("help", false, "")
	root.PersistentFlags().Lookup("help").Hidden = true

	usageTemplate := strings.Replace(root.UsageTemplate(), "\n"+`Use "{{.CommandPath}} [command] --help" for more information about a command.`, "", -1)
	root.SetUsageTemplate(usageTemplate)

	dev := &cobra.Command{
		Use:           "dev",
		Short:         "Start and stop a single vm CF deployment running on your workstation",
		SilenceUsage:  true,
		SilenceErrors: true,
	}
	root.AddCommand(dev)

	dev.AddCommand(NewBosh(Exit, UI, Config))
	dev.AddCommand(NewCatalog(UI, Config))
	dev.AddCommand(NewDownload(Exit, UI, Config))
	dev.AddCommand(NewStart(Exit, UI, Config, Launchd, &process.Manager{}))
	dev.AddCommand(NewStop(Config, Launchd, cfdevdClient.New("CFD3V", Config.CFDevDSocketPath), &process.Manager{}))
	dev.AddCommand(NewTelemetry(UI, Config))
	dev.AddCommand(NewVersion(UI, Config))
	dev.AddCommand(&cobra.Command{
		Use:   "help [command]",
		Short: "Help about any command",
		Run: func(c *cobra.Command, args []string) {
			cmd, _, _ := dev.Find(args)
			cmd.Help()
		},
	})

	return root
}
