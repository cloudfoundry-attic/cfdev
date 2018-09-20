package cmd

import (
	"code.cloudfoundry.org/cfdev/analyticsd/httputil"
	"fmt"
	"gopkg.in/segmentio/analytics-go.v3"
	"net/http"
	"runtime"
	"time"
)

type ServiceBindCmd struct {
	Resource Resource
	IsTimestampSet bool
	Version string
	Uuid string
	EventType string
	T time.Time
	CcHost string
	HttpClient      *http.Client
	AnalyticsClient analytics.Client
}

func(sc *ServiceBindCmd) HandleResponse() error {
	serviceGUID := sc.Resource.Entity.Metadata.Request.Relationships.ServiceInstance.Data.Guid

	var result struct {
		Entity struct {
			ServiceURL string `json:"service_url"`
		}
	}
	serviceEndpoint := "/v2/service_instances/" + serviceGUID
	err := httputil.Fetch(sc.CcHost, serviceEndpoint, sc.Version, sc.Uuid, nil,sc.HttpClient,sc.AnalyticsClient, &result)
	if err != nil {
		return err
	}

	var serviceResponse ServiceResponse
	err = httputil.Fetch(sc.CcHost, result.Entity.ServiceURL, sc.Version, sc.Uuid, nil,sc.HttpClient,sc.AnalyticsClient, &serviceResponse)
	if err != nil {
		return err
	}

	var properties = analytics.Properties{
		"service": serviceResponse.ServiceEntity.ServiceLabel,
		"os":      runtime.GOOS,
		"version": sc.Version,
	}

	if sc.IsTimestampSet {
		err = sc.AnalyticsClient.Enqueue(analytics.Track{
			UserId:     sc.Uuid,
			Event:      sc.EventType,
			Timestamp:  sc.T,
			Properties: properties,
		})
	}

	if err != nil {
		return fmt.Errorf("failed to send analytics: %v", err)
	}

	return nil
}
