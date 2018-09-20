package cmd

import (
	"gopkg.in/segmentio/analytics-go.v3"
	"net/http"
	"time"
)

type ResponseCommand interface {
	HandleResponse() error
}

func CreateHandleResponseCommand(resource Resource, isTimestampSet bool, eventType string, t time.Time, version string, uuid string, ccHost string, httpClient *http.Client,
analyticsClient analytics.Client) ResponseCommand {
	switch eventType {
	case "app created":
		return  &AppCreatedCmd {
			Resource: resource,
			IsTimestampSet: isTimestampSet,
			Version:version,
			Uuid: uuid,
			EventType: eventType,
			T: t,
			CcHost: ccHost,
			HttpClient: httpClient,
			AnalyticsClient:analyticsClient,
		}
	case "service created":
		return &ServiceCreatedCmd {
			Resource: resource,
			IsTimestampSet: isTimestampSet,
			Version: version,
			Uuid: uuid,
			EventType: eventType,
			T: t,
			CcHost: ccHost,
			HttpClient: httpClient,
			AnalyticsClient:analyticsClient,
		}
	case "app bound to service":
		return &ServiceBindCmd{
			Resource: resource,
			IsTimestampSet: isTimestampSet,
			Version: version,
			Uuid: uuid,
			EventType: eventType,
			T: t,
			CcHost: ccHost,
			HttpClient: httpClient,
			AnalyticsClient:analyticsClient,
		}
	}

	return nil
}
