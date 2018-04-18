package cfanalytics

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"runtime"
	"strings"

	"code.cloudfoundry.org/cfdev/config"
	"code.cloudfoundry.org/cfdev/env"
	"github.com/denisbrodbeck/machineid"
	"gopkg.in/segmentio/analytics-go.v3"
)

const START_BEGIN = "start_begin"
const START_END = "start_end"
const STOP = "stop"
const ERROR = "error"

func TrackEvent(event string, data map[string]interface{}, client analytics.Client) error {
	uuid, err := machineid.ProtectedID("cfdev")
	if err != nil {
		uuid = "UNKNOWN_ID"
	}

	var analyticsEvent = &AnalyticsEvent{
		SegClient: client,
		Event:     event,
		UserId:    uuid,
		Data:      data,
		OS:        runtime.GOOS,
		Version:   "0.0.2",
	}

	return analyticsEvent.SendAnalytics()
}

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
