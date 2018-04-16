package cfanalytics

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"

	"strings"

	"code.cloudfoundry.org/cfdev/config"
	"code.cloudfoundry.org/cfdev/env"
)

type UI interface {
	Say(message string, args ...interface{})
	Ask(prompt string) (answer string)
}

func PromptOptIn(conf config.Config, ui UI) error {
	_, err := os.Stat(path.Join(conf.AnalyticsDir, conf.AnalyticsFile))
	if err != nil {
		err = env.SetupAnalytics(conf)
		if err != nil {
			fmt.Printf("ERROR CREATING FILE: %v /n", path.Join(conf.AnalyticsDir, conf.AnalyticsFile))
			return err
		}
	}

	contents, _ := ioutil.ReadFile(path.Join(conf.AnalyticsDir, conf.AnalyticsFile))
	if string(contents[:]) == "" {
		response := ui.Ask(`
CF Dev collects anonymous usage data to help us improve your user experience. We intend to share these anonymous usage analytics with user community by publishing quarterly reports at :

https://github.com/pivotal-cf/cfdev/wiki/Telemetry

Are you ok with CF Dev periodically capturing anonymized telemetry [y/N]?`)
		err = SetTelemetryState(response, conf)
		if err != nil {
			return err
		}
	}

	return nil
}

func SetTelemetryState(response string, conf config.Config) error {
	analyticsFilePath := path.Join(conf.AnalyticsDir, conf.AnalyticsFile)
	_, err := os.Stat(analyticsFilePath)
	if err != nil {
		return err
	}

	if strings.ToLower(response) == "y" || strings.ToLower(response) == "yes" {
		ioutil.WriteFile(analyticsFilePath, []byte("optin"), 0644)
	} else {
		ioutil.WriteFile(analyticsFilePath, []byte("optout"), 0644)
	}

	return nil
}
