package command

import (
	"encoding/json"
	"fmt"
	"gopkg.in/segmentio/analytics-go.v3"
	"log"
	"runtime"
	"time"
)

type ServiceBrokerCreate struct {
	CCClient        CloudControllerClient
	AnalyticsClient analytics.Client
	TimeStamp       time.Time
	UUID            string
	Version         string
	OSVersion       string
	IsBehindProxy   string
	Logger          *log.Logger
}

func (c *ServiceBrokerCreate) HandleResponse(body json.RawMessage) error {
	var properties = analytics.Properties{
		"os":             runtime.GOOS,
		"plugin_version": c.Version,
		"os_version":     c.OSVersion,
		"proxy":          c.IsBehindProxy,
	}

	err := c.AnalyticsClient.Enqueue(analytics.Track{
		UserId:     c.UUID,
		Event:      "created service broker",
		Timestamp:  c.TimeStamp,
		Properties: properties,
	})

	if err != nil {
		return fmt.Errorf("failed to send analytics: %v", err)
	}

	return nil
}
