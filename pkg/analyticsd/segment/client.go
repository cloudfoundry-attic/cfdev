package segment

import (
	"gopkg.in/segmentio/analytics-go.v3"
	"os"
	"runtime"
	"time"
)

//go:generate mockgen -package mocks -destination mocks/analytics.go gopkg.in/segmentio/analytics-go.v3 Client

type Client struct {
	analyticsClient analytics.Client
	uuid            string
	version         string
	osVersion       string
	timeStamp       time.Time
}

func New(analyticsClient analytics.Client, uuid string, version string, osVersion string, timeStamp time.Time) *Client {
	return &Client{
		analyticsClient: analyticsClient,
		version:         version,
		osVersion:       osVersion,
		uuid:            uuid,
		timeStamp:       timeStamp,
	}
}

func (c *Client) Enqueue(event string, properties map[string]string) error {
	isBehindProxy := func() bool {
		return os.Getenv("HTTP_PROXY") != "" ||
			os.Getenv("HTTPS_PROXY") != "" ||
			os.Getenv("http_proxy") != "" ||
			os.Getenv("https_proxy") != ""
	}

	p := analytics.NewProperties()
	p.Set("os", runtime.GOOS)
	p.Set("plugin_version", c.version)
	p.Set("os_version", c.osVersion)
	p.Set("proxy", isBehindProxy())

	for k, v := range properties {
		p.Set(k, v)
	}

	return c.analyticsClient.Enqueue(analytics.Track{
		UserId:     c.uuid,
		Event:      event,
		Timestamp:  c.timeStamp,
		Properties: p,
	})
}
