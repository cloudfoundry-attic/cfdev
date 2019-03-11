package command

import (
	"code.cloudfoundry.org/cfdev/pkg/analyticsd/segment"
	"encoding/json"
	"fmt"
	"log"
)

type AppRestage struct {
	CCClient        CloudControllerClient
	AnalyticsClient *segment.Client
	Logger          *log.Logger
}

func (c *AppRestage) HandleResponse(body json.RawMessage) error {
	err := c.AnalyticsClient.Enqueue("app restage", nil)

	if err != nil {
		return fmt.Errorf("failed to send analytics: %v", err)
	}

	return nil
}
