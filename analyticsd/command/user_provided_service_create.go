package command

import (
	"code.cloudfoundry.org/cfdev/analyticsd/segment"
	"encoding/json"
	"fmt"
	"log"
)

type UserProvidedServiceCreate struct {
	CCClient        CloudControllerClient
	AnalyticsClient *segment.Client
	Logger          *log.Logger
}

func (c *UserProvidedServiceCreate) HandleResponse(body json.RawMessage) error {
	err := c.AnalyticsClient.Enqueue("created user provided service", nil)

	if err != nil {
		return fmt.Errorf("failed to send analytics: %v", err)
	}

	return nil
}
