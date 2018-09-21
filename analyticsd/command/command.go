package command

import (
	"encoding/json"
	"gopkg.in/segmentio/analytics-go.v3"
	"log"
	"net/url"
	"time"
)

//go:generate mockgen -package mocks -destination mocks/analytics.go gopkg.in/segmentio/analytics-go.v3 Client

type Command interface {
	HandleResponse(body json.RawMessage) error
}

//go:generate mockgen -package mocks -destination mocks/cloud_controller_client.go code.cloudfoundry.org/cfdev/analyticsd/command CloudControllerClient
type CloudControllerClient interface {
	Fetch(path string, params url.Values, dest interface{}) error
}

func New(event string,
	ccClient CloudControllerClient,
	analyticsClient analytics.Client,
	timeStamp time.Time,
	UUID string,
	version string,
	logger *log.Logger) (Command, bool) {

	switch event {
	case "audit.app.create":
		logger.Printf("Detected event for %q\n", event)

		return &AppCreate{
			CCclient: ccClient,
			AnalyticsClient: analyticsClient,
			TimeStamp: timeStamp,
			UUID: UUID,
			Version: version,
			Logger: logger,
		}, true
	case "audit.service_instance.create":
		logger.Printf("Detected event for %q\n", event)

		return &ServiceCreate{
			CCclient: ccClient,
			AnalyticsClient: analyticsClient,
			TimeStamp: timeStamp,
			UUID: UUID,
			Version: version,
			Logger: logger,
		}, true
	case "audit.service_binding.create":
		logger.Printf("Detected event for %q\n", event)

		return &ServiceBind{
			CCclient: ccClient,
			AnalyticsClient: analyticsClient,
			TimeStamp: timeStamp,
			UUID: UUID,
			Version: version,
			Logger: logger,
		}, true
	default:
		return nil, false
	}
}