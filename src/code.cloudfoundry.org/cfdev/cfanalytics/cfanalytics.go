package cfanalytics

import (
	"io/ioutil"
	"os"
	"path"
	"runtime"
	"strings"

	"code.cloudfoundry.org/cfdev/config"
	"github.com/denisbrodbeck/machineid"
	"gopkg.in/segmentio/analytics-go.v3"
)

const START_BEGIN = "start_begin"
const START_END = "start_end"
const STOP = "stop"
const ERROR = "error"
const UNINSTALL = "uninstall"

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
	contents, _ := ioutil.ReadFile(path.Join(conf.AnalyticsDir, conf.AnalyticsFile))
	if string(contents) == "" {
		response := ui.Ask(`
CF Dev collects anonymous usage data to help us improve your user experience. We intend to share these anonymous usage analytics with user community by publishing quarterly reports at :

https://github.com/pivotal-cf/cfdev/wiki/Telemetry

Are you ok with CF Dev periodically capturing anonymized telemetry [y/N]?`)
		if err := SetTelemetryState(response, conf); err != nil {
			return err
		}
	}

	return nil
}

func SetTelemetryState(response string, conf config.Config) error {
	if err := os.MkdirAll(conf.AnalyticsDir, 0755); err != nil {
		return err
	}

	fileContents := "optout"
	if strings.ToLower(response) == "y" || strings.ToLower(response) == "yes" {
		fileContents = "optin"
	}

	return ioutil.WriteFile(path.Join(conf.AnalyticsDir, conf.AnalyticsFile), []byte(fileContents), 0644)
}
