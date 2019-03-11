package command

import (
	"code.cloudfoundry.org/cfdev/pkg/analyticsd/segment"
	"encoding/json"
	"fmt"
	"log"
)

type AppCrash struct {
	CCClient        CloudControllerClient
	AnalyticsClient *segment.Client
	Logger          *log.Logger
}

func (c *AppCrash) HandleResponse(body json.RawMessage) error {
	err := c.AnalyticsClient.Enqueue("app push failed", nil)

	if err != nil {
		return fmt.Errorf("failed to send analytics: %v", err)
	}

	return nil
}
