package command

import (
	"encoding/json"
	"fmt"
	"gopkg.in/segmentio/analytics-go.v3"
	"log"
	"runtime"
	"time"
)

type UserProvidedServiceCreate struct {
	CCClient        CloudControllerClient
	AnalyticsClient analytics.Client
	TimeStamp       time.Time
	UUID            string
	Version         string
	Logger          *log.Logger
}

func (c *UserProvidedServiceCreate) HandleResponse(body json.RawMessage) error {
	var properties = analytics.Properties{
		"os":        runtime.GOOS,
		"version":   c.Version,
	}

	err := c.AnalyticsClient.Enqueue(analytics.Track{
		UserId:     c.UUID,
		Event:      "created user provided service",
		Timestamp:  c.TimeStamp,
		Properties: properties,
	})

	if err != nil {
		return fmt.Errorf("failed to send analytics: %v", err)
	}

	return nil
}
