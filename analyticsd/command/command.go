package command

import (
	"code.cloudfoundry.org/cfdev/analyticsd/config"
	"encoding/json"
	"gopkg.in/segmentio/analytics-go.v3"
	"log"
	"net/url"
	"strings"
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

func New(
	event string,
	ccClient CloudControllerClient,
	analyticsClient analytics.Client,
	timeStamp time.Time,
	UUID string,
	version string,
	logger *log.Logger) (Command, bool) {

	switch event {
	case "audit.app.restage":
		logger.Printf("Detected event for %q\n", event)

		return &AppRestage{
			CCClient:        ccClient,
			AnalyticsClient: analyticsClient,
			TimeStamp:       timeStamp,
			UUID:            UUID,
			Version:         version,
			Logger:          logger,
		}, true
	case "audit.app.create":
		logger.Printf("Detected event for %q\n", event)

		return &AppCreate{
			CCClient:        ccClient,
			AnalyticsClient: analyticsClient,
			TimeStamp:       timeStamp,
			UUID:            UUID,
			Version:         version,
			Logger:          logger,
		}, true
	case "app.crash":
		logger.Printf("Detected event for %q\n", event)

		return &AppCrash{
			CCClient:        ccClient,
			AnalyticsClient: analyticsClient,
			TimeStamp:       timeStamp,
			UUID:            UUID,
			Version:         version,
			Logger:          logger,
		}, true
	case "audit.service_instance.create":
		logger.Printf("Detected event for %q\n", event)

		return &ServiceCreate{
			CCClient:        ccClient,
			AnalyticsClient: analyticsClient,
			TimeStamp:       timeStamp,
			UUID:            UUID,
			Version:         version,
			Logger:          logger,
		}, true
	case "audit.service_binding.create":
		logger.Printf("Detected event for %q\n", event)

		return &ServiceBind{
			CCClient:        ccClient,
			AnalyticsClient: analyticsClient,
			TimeStamp:       timeStamp,
			UUID:            UUID,
			Version:         version,
			Logger:          logger,
		}, true
	case "audit.user_provided_service_instance.create":
		logger.Printf("Detected event for %q\n", event)

		return &UserProvidedServiceCreate{
			CCClient:        ccClient,
			AnalyticsClient: analyticsClient,
			TimeStamp:       timeStamp,
			UUID:            UUID,
			Version:         version,
			Logger:          logger,
		}, true
	default:
		return nil, false
	}
}

func serviceIsWhiteListed(serviceLabel string) bool {
	for _, listedLabel := range config.SERVICE_WHITELIST {
		sl, ll := strings.ToLower(serviceLabel), strings.ToLower(listedLabel)
		if sl == ll {
			return true
		}
	}

	return false
}
