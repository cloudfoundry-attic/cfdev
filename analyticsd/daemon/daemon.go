package daemon

//go:generate mockgen -package mocks -destination mocks/analytics.go gopkg.in/segmentio/analytics-go.v3 Client

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"runtime"
	"strings"
	"time"

	"gopkg.in/segmentio/analytics-go.v3"
)

const ccTimeStampFormat = "2006-01-02T15:04:05Z"

type Daemon struct {
	ccHost          string
	httpClient      *http.Client
	UUID            string
	analyticsClient analytics.Client
	ticker          *time.Ticker
	pollingInterval time.Duration
	logger          *log.Logger
	lastTime        *time.Time
	doneChan        chan bool
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
}

func New(
	ccHost string,
	UUID string,
	writer io.Writer,
	httpClient *http.Client,
	analyticsClient analytics.Client,
	pollingInterval time.Duration,
) *Daemon {
	return &Daemon{
		ccHost:          ccHost,
		UUID:            UUID,
		httpClient:      httpClient,
		analyticsClient: analyticsClient,
		ticker:          time.NewTicker(pollingInterval),
		pollingInterval: pollingInterval,
		logger:          log.New(writer, "[ANALYTICSD] ", log.LstdFlags),
		doneChan:        make(chan bool, 1),
	}
}

type Request struct {
	Buildpack string
}

type Metadata struct {
	Request Request
}

type Entity struct {
	Type      string
	Timestamp string
	Metadata  Metadata
}

type Resource struct {
	Entity Entity
}

type Response struct {
	Resources []Resource
}

var (
	eventTypes = map[string]string{
		"audit.app.create": "app push",
	}
)

func (d *Daemon) Start() {
	err := d.do()
	if err != nil {
		d.logger.Println(err)
	}
	for {
		select {
		case <-d.doneChan:
			return
		case <-time.NewTicker(d.pollingInterval).C:
			err := d.do()

			if err != nil {
				d.logger.Println(err)
			}
		}
	}
}

func (d *Daemon) Stop() {
	d.doneChan <- true
}

func (d *Daemon) do() error {
	req, err := http.NewRequest(http.MethodGet, d.ccHost+"/v2/events", nil)
	if err != nil {
		return err
	}

	params := url.Values{}
	params.Add("q", "type IN "+eventTypesFilter())

	lastTimeIsSet := d.lastTime != nil

	if lastTimeIsSet {
		params.Add("q", "timestamp>"+d.lastTime.Format(ccTimeStampFormat))
	}

	req.URL.RawQuery = params.Encode()

	resp, err := d.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to query cloud controller: %s", err)
	}

	contents, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		var properties = analytics.Properties{
			"message": fmt.Sprintf("failed to contact cc api: [%v] %s", resp.Status, contents),
			"os":      runtime.GOOS,
		}

		err := d.analyticsClient.Enqueue(analytics.Track{
			UserId:     d.UUID,
			Event:      "analytics error",
			Timestamp:  time.Now().UTC(),
			Properties: properties,
		})

		if err != nil {
			return fmt.Errorf("failed to send analytics: %v", err)
		}

		return nil
	}

	var appResponse Response

	err = json.Unmarshal(contents, &appResponse)
	if err != nil {
		return err
	}

	if len(appResponse.Resources) == 0 {
		d.saveLatestTime(time.Now())
	}

	for _, resource := range appResponse.Resources {
		eventType, ok := eventTypes[resource.Entity.Type]
		if !ok {
			continue
		}

		t, err := time.Parse(time.RFC3339, resource.Entity.Timestamp)
		if err != nil {
			return err
		}

		d.saveLatestTime(t)

		buildpack, ok := buildpackWhitelist[resource.Entity.Metadata.Request.Buildpack]
		if !ok {
			buildpack = "custom"
		}
		var properties = analytics.Properties{
			"buildpack": buildpack,
			"os":        runtime.GOOS,
		}

		if lastTimeIsSet {
			err = d.analyticsClient.Enqueue(analytics.Track{
				UserId:     d.UUID,
				Event:      eventType,
				Timestamp:  t,
				Properties: properties,
			})
		}

		if err != nil {
			return fmt.Errorf("failed to send analytics: %v", err)
		}
	}

	return nil
}

func eventTypesFilter() string {
	var coll []string
	for k, _ := range eventTypes {
		coll = append(coll, k)
	}
	return strings.Join(coll, ",")
}

func (d *Daemon) saveLatestTime(t time.Time) {
	t = t.UTC()
	if d.lastTime == nil || t.After(*d.lastTime) {
		d.lastTime = &t
	}
}
