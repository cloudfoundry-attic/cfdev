package command

import (
	"encoding/json"
	"fmt"
	"gopkg.in/segmentio/analytics-go.v3"
	"log"
	"runtime"
	"time"
)

type ServiceBind struct {
	CCClient        CloudControllerClient
	AnalyticsClient analytics.Client
	TimeStamp       time.Time
	UUID            string
	Version         string
	OSVersion       string
	IsBehindProxy   string
	Logger          *log.Logger
}

func (c *ServiceBind) HandleResponse(body json.RawMessage) error {
	var metadata struct {
		Request struct {
			Relationships struct {
				ServiceInstance struct {
					Data struct {
						Guid string
					}
				} `json:"service_instance"`
			}
		}
	}

	json.Unmarshal(body, &metadata)

	var urlResp struct {
		Entity struct {
			ServiceURL string `json:"service_url"`
		}
	}

	path := "/v2/service_instances/" + metadata.Request.Relationships.ServiceInstance.Data.Guid
	err := c.CCClient.Fetch(path, nil, &urlResp)
	if err != nil {
		return fmt.Errorf("failed to make request to: %s: %s", path, err)
	}

	var labelResp struct {
		Entity struct {
			Label string
		}
	}

	path = urlResp.Entity.ServiceURL
	err = c.CCClient.Fetch(path, nil, &labelResp)
	if err != nil {
		return fmt.Errorf("failed to make request to: %s: %s", path, err)
	}

	if !serviceIsWhiteListed(labelResp.Entity.Label) {
		return nil
	}

	var properties = analytics.Properties{
		"service":        labelResp.Entity.Label,
		"os":             runtime.GOOS,
		"plugin_version": c.Version,
		"os_version":     c.OSVersion,
		"proxy":          c.IsBehindProxy,
	}

	err = c.AnalyticsClient.Enqueue(analytics.Track{
		UserId:     c.UUID,
		Event:      "app bound to service",
		Timestamp:  c.TimeStamp,
		Properties: properties,
	})

	if err != nil {
		return fmt.Errorf("failed to send analytics: %v", err)
	}

	return nil
}
