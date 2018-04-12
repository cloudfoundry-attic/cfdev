package cmd

import (
	"io/ioutil"
	"path"

	"strings"

	"code.cloudfoundry.org/cfdev/cfanalytics"
	"code.cloudfoundry.org/cfdev/config"
)

type Telemetry struct {
	Exit   chan struct{}
	UI     UI
	Config config.Config
}

func (t *Telemetry) Run(args []string) error {
	analyticsPath := path.Join(t.Config.AnalyticsDir, t.Config.AnalyticsFile)

	if len(args) == 0 {
		contents, err := ioutil.ReadFile(analyticsPath)
		if err != nil {
			return err
		}

		if string(contents[:]) == "optin" {
			t.UI.Say("Telemetry is turned ON")
		} else {
			t.UI.Say("Telemetry is turned OFF")
		}
	} else if strings.ToLower(args[0]) == "on" {
		err := cfanalytics.SetTelemetryState("yes", t.Config)
		if err != nil {
			return err
		}

		t.UI.Say("Telemetry is turned ON")
	} else if strings.ToLower(args[0]) == "off" {
		err := cfanalytics.SetTelemetryState("no", t.Config)
		if err != nil {
			return err
		}

		t.UI.Say("Telemetry is turned OFF")
	}

	return nil
}
