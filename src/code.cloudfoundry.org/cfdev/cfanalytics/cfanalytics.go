package cfanalytics

import (
	"runtime"
	"strings"
	"time"

	"code.cloudfoundry.org/cfdev/errors"
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
	GetProps() map[string]interface{}
}
type UI interface {
	Ask(prompt string) (answer string)
}

type Analytics struct {
	client  analytics.Client
	toggle  Toggle
	userId  string
	version string
	exit    chan struct{}
	ui      UI
}

func New(toggle Toggle, client analytics.Client, version string, exit chan struct{}, ui UI) *Analytics {
	uuid, err := machineid.ProtectedID("cfdev")
	if err != nil {
		uuid = "UNKNOWN_ID"
	}
	return &Analytics{
		client:  client,
		toggle:  toggle,
		userId:  uuid,
		version: version,
		exit:    exit,
		ui:      ui,
	}
}

func (a *Analytics) Close() {
	a.client.Close()
}

func (a *Analytics) Event(event string, data ...map[string]interface{}) error {
	if !a.toggle.Get() {
		return nil
	}
	properties := analytics.NewProperties()
	properties.Set("os", runtime.GOOS)
	properties.Set("version", a.version)
	for k, v := range a.toggle.GetProps() {
		properties.Set(k, v)
	}
	for _, d := range data {
		for k, v := range d {
			properties.Set(k, v)
		}
	}

	return a.client.Enqueue(analytics.Track{
		UserId:     a.userId,
		Event:      event,
		Timestamp:  time.Now().UTC(),
		Properties: properties,
	})
}

func (a *Analytics) PromptOptIn() error {
	if !a.toggle.Defined() {
		response := a.ui.Ask(`
CF Dev collects anonymous usage data to help us improve your user experience. We intend to share these anonymous usage analytics with user community by publishing quarterly reports at :

https://github.com/pivotal-cf/cfdev/wiki/Telemetry

Are you ok with CF Dev periodically capturing anonymized telemetry [y/N]?`)

		select {
		case <-a.exit:
			return errors.SafeWrap(nil, "Exit while waiting for telemetry prompt")
		case <-time.After(time.Millisecond):
		}

		response = strings.ToLower(response)
		enabled := response == "y" || response == "yes"
		if err := a.toggle.Set(enabled); err != nil {
			return err
		}
	}
	return nil
}
