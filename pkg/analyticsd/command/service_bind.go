package command

import (
	"code.cloudfoundry.org/cfdev/pkg/analyticsd/segment"
	"encoding/json"
	"fmt"
	"log"
)

type ServiceBind struct {
	CCClient        CloudControllerClient
	AnalyticsClient *segment.Client
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

	err = c.AnalyticsClient.Enqueue("app bound to service", map[string]string{
		"service": labelResp.Entity.Label,
	})

	if err != nil {
		return fmt.Errorf("failed to send analytics: %v", err)
	}

	return nil
}
