package cmd

import (
	"fmt"
	"gopkg.in/segmentio/analytics-go.v3"
	"net/http"
	"runtime"
	"time"
)

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

type AppCreatedCmd struct {
	Resource Resource
	IsTimestampSet bool
	Version string
	Uuid string
	EventType string
	T time.Time
	CcHost string
	HttpClient      *http.Client
	AnalyticsClient analytics.Client
}

func(ac *AppCreatedCmd) HandleResponse() error {
	buildpack, ok := buildpackWhitelist[ac.Resource.Entity.Metadata.Request.Buildpack]
	if !ok {
		buildpack = "custom"
	}
	var properties = analytics.Properties{
		"buildpack": buildpack,
		"os":        runtime.GOOS,
		"version":   ac.Version,
	}

	var err error

	if ac.IsTimestampSet {
		err = ac.AnalyticsClient.Enqueue(analytics.Track{
			UserId:     ac.Uuid,
			Event:      ac.EventType,
			Timestamp:  ac.T,
			Properties: properties,
		})
	}

	if err != nil {
		return fmt.Errorf("failed to send analytics: %v", err)
	}

	return nil
}
