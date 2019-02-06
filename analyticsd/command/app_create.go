package command

import (
	"encoding/json"
	"fmt"
	"gopkg.in/segmentio/analytics-go.v3"
	"log"
	"runtime"
	"time"
)

type AppCreate struct {
	CCClient        CloudControllerClient
	AnalyticsClient analytics.Client
	TimeStamp       time.Time
	UUID            string
	Version         string
	OSVersion       string
	IsBehindProxy   string
	Logger          *log.Logger
}

var buildpackWhitelist = map[string]string{
	"staticfile_buildpack":  "staticfile",
	"java_buildpack":        "java",
	"ruby_buildpack":        "ruby",
	"dotnet_core_buildpack": "dotnet_core",
	"nodejs_buildpack":      "nodejs",
	"go_buildpack":          "go",
	"python_buildpack":      "python",
	"php_buildpack":         "php",
	"binary_buildpack":      "binary",
	"":                      "unspecified",
}

func (c *AppCreate) HandleResponse(body json.RawMessage) error {
	var metadata struct {
		Request struct {
			Buildpack string
		}
	}

	json.Unmarshal(body, &metadata)

	buildpack, ok := buildpackWhitelist[metadata.Request.Buildpack]
	if !ok {
		buildpack = "custom"
	}

	var properties = analytics.Properties{
		"buildpack":      buildpack,
		"os":             runtime.GOOS,
		"plugin_version": c.Version,
		"os_version":     c.OSVersion,
		"proxy":          c.IsBehindProxy,
	}

	err := c.AnalyticsClient.Enqueue(analytics.Track{
		UserId:     c.UUID,
		Event:      "app created",
		Timestamp:  c.TimeStamp,
		Properties: properties,
	})

	if err != nil {
		return fmt.Errorf("failed to send analytics: %v", err)
	}

	return nil
}
