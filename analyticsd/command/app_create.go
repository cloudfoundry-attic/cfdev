package command

import (
	"code.cloudfoundry.org/cfdev/analyticsd/segment"
	"encoding/json"
	"fmt"
	"log"
)

type AppCreate struct {
	CCClient        CloudControllerClient
	AnalyticsClient *segment.Client
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

	err := c.AnalyticsClient.Enqueue("app created", map[string]string{
		"buildpack": buildpack,
	})

	if err != nil {
		return fmt.Errorf("failed to send analytics: %v", err)
	}

	return nil
}
