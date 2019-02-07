package command

import (
	"code.cloudfoundry.org/cfdev/analyticsd/config"
	"code.cloudfoundry.org/cfdev/analyticsd/segment"
	"encoding/json"
	"gopkg.in/segmentio/analytics-go.v3"
	"log"
	"net/url"
	"strings"
	"time"
)

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
	osVersion string,
	logger *log.Logger) (Command, bool) {

	logger.Printf("Detected event for %q\n", event)

	segmentClient := segment.New(
		analyticsClient,
		UUID,
		version,
		osVersion,
		timeStamp,
	)

	switch event {
	case "audit.app.restage":
		return &AppRestage{
			CCClient:        ccClient,
			AnalyticsClient: segmentClient,
			Logger:          logger,
		}, true
	case "audit.app.create":
		return &AppCreate{
			CCClient:        ccClient,
			AnalyticsClient: segmentClient,
			Logger:          logger,
		}, true
	case "app.crash":
		return &AppCrash{
			CCClient:        ccClient,
			AnalyticsClient: segmentClient,
			Logger:          logger,
		}, true
	case "audit.organization.create":
		return &OrgCreate{
			CCClient:        ccClient,
			AnalyticsClient: segmentClient,
			Logger:          logger,
		}, true
	case "audit.space.create":
		return &SpaceCreate{
			CCClient:        ccClient,
			AnalyticsClient: segmentClient,
			Logger:          logger,
		}, true
	case "audit.service_instance.create":
		return &ServiceCreate{
			CCClient:        ccClient,
			AnalyticsClient: segmentClient,
			Logger:          logger,
		}, true
	case "audit.service_binding.create":
		return &ServiceBind{
			CCClient:        ccClient,
			AnalyticsClient: segmentClient,
			Logger:          logger,
		}, true
	case "audit.service_broker.create":
		return &ServiceBrokerCreate{
			CCClient:        ccClient,
			AnalyticsClient: segmentClient,
			Logger:          logger,
		}, true
	case "audit.user_provided_service_instance.create":
		return &UserProvidedServiceCreate{
			CCClient:        ccClient,
			AnalyticsClient: segmentClient,
			Logger:          logger,
		}, true
	case "audit.route.create":
		return &RouteCreate{
			CCClient:        ccClient,
			AnalyticsClient: segmentClient,
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
