package cmd

import (
	"io/ioutil"
	"path"

	"code.cloudfoundry.org/cfdev/cfanalytics"
	"code.cloudfoundry.org/cfdev/config"
	"github.com/spf13/cobra"
)

func NewTelemetry(UI UI, Config config.Config) *cobra.Command {
	var flagOff, flagOn bool
	cmd := &cobra.Command{
		Use:   "telemetry",
		Short: "Show status for collecting anonymous usage telemetry",
		RunE: func(cmd *cobra.Command, args []string) error {
			analyticsPath := path.Join(Config.AnalyticsDir, Config.AnalyticsFile)
			if flagOff {
				if err := cfanalytics.SetTelemetryState("no", Config); err != nil {
					return err
				}
			} else if flagOn {
				if err := cfanalytics.SetTelemetryState("yes", Config); err != nil {
					return err
				}
			}

			contents, err := ioutil.ReadFile(analyticsPath)
			if err != nil {
				return err
			}

			if string(contents) == "optin" {
				UI.Say("Telemetry is turned ON")
			} else {
				UI.Say("Telemetry is turned OFF")
			}
			return nil
		},
	}

	cmd.PersistentFlags().BoolVar(&flagOff, "off", false, "Disable the collection of anonymous usage telemetryDisable the collection of anonymous usage telemetry")
	cmd.PersistentFlags().BoolVar(&flagOn, "on", false, "Enable the collection of anonymous usage telemetryDisable the collection of anonymous usage telemetry")
	return cmd
}
