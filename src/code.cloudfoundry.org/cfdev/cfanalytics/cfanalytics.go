package cfanalytics

import (
	"runtime"
	"strings"
	"time"

	"github.com/denisbrodbeck/machineid"
	analytics "gopkg.in/segmentio/analytics-go.v3"
)

const START_BEGIN = "start_begin"
const START_END = "start_end"
const STOP = "stop"
const ERROR = "error"
const UNINSTALL = "uninstall"

type Toggle interface {
	Defined() bool
	Get() bool
	Set(value bool) error
}

type Analytics struct {
	client  analytics.Client
	toggle  Toggle
	userId  string
	version string
}

func New(toggle Toggle, client analytics.Client) *Analytics {
	uuid, err := machineid.ProtectedID("cfdev")
	if err != nil {
		uuid = "UNKNOWN_ID"
	}
	return &Analytics{
		client:  client,
		toggle:  toggle,
		userId:  uuid,
		version: "0.0.2",
	}
}

func (a *Analytics) Close() {
	a.client.Close()
}

func (a *Analytics) Event(event string, data map[string]interface{}) error {
	if !a.toggle.Get() {
		return nil
	}
	properties := analytics.NewProperties()
	properties.Set("os", runtime.GOOS)
	properties.Set("version", a.version)
	properties.Set("localtime", time.Now())
	for k, v := range data {
		properties.Set(k, v)
	}

	return a.client.Enqueue(analytics.Track{
		UserId:     a.userId,
		Event:      event,
		Properties: properties,
	})
}

type UI interface {
	Ask(prompt string) (answer string)
}

func (a *Analytics) PromptOptIn(ui UI) error {
	if !a.toggle.Defined() {
		response := ui.Ask(`
CF Dev collects anonymous usage data to help us improve your user experience. We intend to share these anonymous usage analytics with user community by publishing quarterly reports at :

https://github.com/pivotal-cf/cfdev/wiki/Telemetry

Are you ok with CF Dev periodically capturing anonymized telemetry [y/N]?`)

		response = strings.ToLower(response)
		enabled := response == "y" || response == "yes"
		if err := a.toggle.Set(enabled); err != nil {
			return err
		}
	}
	return nil
}
