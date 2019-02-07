package command

import (
	"code.cloudfoundry.org/cfdev/analyticsd/segment"
	"encoding/json"
	"fmt"
	"log"
)

type OrgCreate struct {
	CCClient        CloudControllerClient
	AnalyticsClient *segment.Client
	Logger          *log.Logger
}

func (c *OrgCreate) HandleResponse(body json.RawMessage) error {
	err := c.AnalyticsClient.Enqueue("org created", nil)

	if err != nil {
		return fmt.Errorf("failed to send analytics: %v", err)
	}

	return nil
}
