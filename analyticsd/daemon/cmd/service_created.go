package cmd

import (
	"code.cloudfoundry.org/cfdev/analyticsd/httputil"
	"fmt"
	"gopkg.in/segmentio/analytics-go.v3"
	"net/http"
	"runtime"
	"time"
)

type Request struct {
	Buildpack       string
	ServicePlanGUID string `json:"service_plan_guid"`
	Relationships Relationships
}

type Relationships struct {
	ServiceInstance ServiceInstance `json:"service_instance"`
}

type Metadata struct {
	Request Request
}

type Entity struct {
	Type      string
	Timestamp string
	Metadata  Metadata
}

type Resource struct {
	Entity Entity
}

type Response struct {
	NextURL   *string `json:"next_url"`
	Resources []Resource
}

type ServicePlanResponse struct {
	ServicePlanEntity ServicePlanEntity `json:"entity"`
}

type ServicePlanEntity struct {
	ServicePlanGUID string `json:"service_guid"`
}

type ServiceResponse struct {
	ServiceEntity ServiceEntity `json:"entity"`
}

type ServiceEntity struct {
	ServiceLabel string `json:"label"`
}

type ServiceInstance struct {
	Data Data `json:"data"`
}

type Data struct {
	Guid string `json:"guid"`
}

type ServiceCreatedCmd struct {
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

func(sc *ServiceCreatedCmd) HandleResponse() error {
	var servicePlanResponse ServicePlanResponse
	servicePlanEndpoint := "/v2/service_plans/" + sc.Resource.Entity.Metadata.Request.ServicePlanGUID
	err := httputil.Fetch(sc.CcHost, servicePlanEndpoint, sc.Version, sc.Uuid, nil,sc.HttpClient,sc.AnalyticsClient, &servicePlanResponse)
	if err != nil {
		return err
	}

	serviceGUID := servicePlanResponse.ServicePlanEntity.ServicePlanGUID

	var serviceResponse ServiceResponse
	serviceEndpoint := "/v2/services/" + serviceGUID
	err = httputil.Fetch(sc.CcHost, serviceEndpoint, sc.Version, sc.Uuid, nil,sc.HttpClient,sc.AnalyticsClient, &serviceResponse)
	if err != nil {
		return err
	}
	serviceType := serviceResponse.ServiceEntity.ServiceLabel

	var properties = analytics.Properties{
		"service": serviceType,
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
