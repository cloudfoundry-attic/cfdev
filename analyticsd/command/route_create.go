package command

import (
	"code.cloudfoundry.org/cfdev/analyticsd/segment"
	"encoding/json"
	"fmt"
	"log"
)

type RouteCreate struct {
	CCClient        CloudControllerClient
	AnalyticsClient *segment.Client
	Logger          *log.Logger
}

func (c *RouteCreate) HandleResponse(body json.RawMessage) error {
	err := c.AnalyticsClient.Enqueue("created route", nil)

	if err != nil {
		return fmt.Errorf("failed to send analytics: %v", err)
	}

	return nil
}
