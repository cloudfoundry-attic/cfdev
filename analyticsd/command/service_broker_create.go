package command

import (
	"code.cloudfoundry.org/cfdev/analyticsd/segment"
	"encoding/json"
	"fmt"
	"log"
)

type ServiceBrokerCreate struct {
	CCClient        CloudControllerClient
	AnalyticsClient *segment.Client
	Logger          *log.Logger
}

func (c *ServiceBrokerCreate) HandleResponse(body json.RawMessage) error {
	err := c.AnalyticsClient.Enqueue( "created service broker", nil)

	if err != nil {
		return fmt.Errorf("failed to send analytics: %v", err)
	}

	return nil
}
