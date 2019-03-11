package command

import (
	"code.cloudfoundry.org/cfdev/pkg/analyticsd/segment"
	"encoding/json"
	"fmt"
	"log"
)

type SpaceCreate struct {
	CCClient        CloudControllerClient
	AnalyticsClient *segment.Client
	Logger          *log.Logger
}

func (c *SpaceCreate) HandleResponse(body json.RawMessage) error {
	err := c.AnalyticsClient.Enqueue("space created", nil)

	if err != nil {
		return fmt.Errorf("failed to send analytics: %v", err)
	}

	return nil
}
