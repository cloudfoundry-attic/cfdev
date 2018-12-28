package cfanalytics

import (
	"code.cloudfoundry.org/cfdev/errors"
	"runtime"
	"strings"
	"time"

	"github.com/denisbrodbeck/machineid"
	"gopkg.in/segmentio/analytics-go.v3"
)

const (
	START_BEGIN      = "start_begin"
	START_END        = "start_end"
	SELECTED_SERVICE = "selected_service"
	STOP             = "stop"
	STOP_TELEMETRY   = "telemetry off"
	BOSH_ENV         = "bosh"
	ERROR            = "error"
	UNINSTALL        = "uninstall"
	DEPLOY_SERVICE   = "deployed service"
)

//go:generate mockgen -package mocks -destination mocks/analytics_client.go gopkg.in/segmentio/analytics-go.v3 Client

//go:generate mockgen -package mocks -destination mocks/toggle.go code.cloudfoundry.org/cfdev/cfanalytics Toggle
type Toggle interface {
	Defined() bool
	CustomAnalyticsDefined() bool
	Enabled() bool
	IsCustom() bool
	SetCFAnalyticsEnabled(value bool) error
	SetCustomAnalyticsEnabled(value bool) error
	GetProps() map[string]interface{}
	SetProp(k, v string) error
}

//go:generate mockgen -package mocks -destination mocks/ui.go code.cloudfoundry.org/cfdev/cfanalytics UI
type UI interface {
	Ask(prompt string) (answer string)
}

type Analytics struct {
	client    analytics.Client
	toggle    Toggle
	userId    string
	version   string
	osVersion string
	exit      chan struct{}
	ui        UI
}

func New(toggle Toggle, client analytics.Client, version string, osVersion string, exit chan struct{}, ui UI) *Analytics {
	uuid, err := machineid.ProtectedID("cfdev")
	if err != nil {
		uuid = "UNKNOWN_ID"
	}

	return &Analytics{
		client:    client,
		toggle:    toggle,
		userId:    uuid,
		version:   version,
		osVersion: osVersion,
		exit:      exit,
		ui:        ui,
	}
}

func (a *Analytics) Close() {
	a.client.Close()
}

func (a *Analytics) Event(event string, data ...map[string]interface{}) error {
	if !a.toggle.Enabled() {
		return nil
	}

	a.client.Enqueue(analytics.Identify{
		UserId: a.userId,
	})

	properties := analytics.NewProperties()
	properties.Set("os", runtime.GOOS)
	properties.Set("plugin_version", a.version)
	properties.Set("os_version", a.osVersion)
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

func (a *Analytics) PromptOptInIfNeeded(customMessage string) error {
	useCustom := customMessage != ""

	if !a.toggle.Defined() || (useCustom && !a.toggle.CustomAnalyticsDefined()) {

		message := `CF Dev collects anonymous usage data to help us improve your user experience. We intend to share these anonymous usage analytics with user community by publishing quarterly reports at :
		
https://github.com/pivotal-cf/cfdev/wiki/Telemetry
		
Are you ok with CF Dev periodically capturing anonymized telemetry [y/N]?`
		if useCustom {
			message = customMessage
		}
		response := a.ui.Ask(message)

		select {
		case <-a.exit:
			return errors.SafeWrap(nil, "Exit while waiting for telemetry prompt")
		case <-time.After(time.Millisecond):
		}

		response = strings.ToLower(response)
		enabled := response == "y" || response == "yes"

		if useCustom {
			if err := a.toggle.SetCustomAnalyticsEnabled(enabled); err != nil {
				return err
			}
		} else {
			if err := a.toggle.SetCFAnalyticsEnabled(enabled); err != nil {
				return err
			}
		}
	}
	return nil
}
