package daemon

//go:generate mockgen -package mocks -destination mocks/analytics.go gopkg.in/segmentio/analytics-go.v3 Client

import (
	"code.cloudfoundry.org/cfdev/analyticsd/daemon/cmd"
	"code.cloudfoundry.org/cfdev/analyticsd/httputil"
	"fmt"
	"gopkg.in/segmentio/analytics-go.v3"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const ccTimeStampFormat = "2006-01-02T15:04:05Z"

type Daemon struct {
	CcHost          string
	httpClient      *http.Client
	UUID            string
	version         string
	analyticsClient analytics.Client
	ticker          *time.Ticker
	pollingInterval time.Duration
	logger          *log.Logger
	lastTime        *time.Time
	doneChan        chan bool
}

func New(
	ccHost string,
	UUID string,
	version string,
	writer io.Writer,
	httpClient *http.Client,
	analyticsClient analytics.Client,
	pollingInterval time.Duration,
) *Daemon {
	return &Daemon{
		CcHost:          ccHost,
		UUID:            UUID,
		version:         version,
		httpClient:      httpClient,
		analyticsClient: analyticsClient,
		ticker:          time.NewTicker(pollingInterval),
		pollingInterval: pollingInterval,
		logger:          log.New(writer, "[ANALYTICSD] ", log.LstdFlags),
		doneChan:        make(chan bool, 1),
	}
}

var (
	eventTypes = map[string]string{
		"audit.app.create":              "app created",
		"audit.service_instance.create": "service created",
	}
)

func (d *Daemon) Start() {
	err := d.do(false)
	if err != nil {
		d.logger.Println(err)
	}
	for {
		select {
		case <-d.doneChan:
			return
		case <-time.NewTicker(d.pollingInterval).C:
			isTimestampSet := d.lastTime != nil
			err := d.do(isTimestampSet)

			if err != nil {
				d.logger.Println(err)
			}
		}
	}
}

func (d *Daemon) Stop() {
	d.doneChan <- true
}

func (d *Daemon) do(isTimestampSet bool) error {
	var (
		nextURL   *string = nil
		resources []cmd.Resource
		fetch     = func(params url.Values) error {
			var appResponse cmd.Response
			err := httputil.Fetch(d.CcHost, "/v2/events", d.version, d.UUID, params, d.httpClient, d.analyticsClient, &appResponse)
			if err != nil {
				return err
			}

			resources = append(resources, appResponse.Resources...)
			nextURL = appResponse.NextURL
			return nil
		}
	)

	params := url.Values{}
	params.Add("q", "type IN "+eventTypesFilter())
	if isTimestampSet {
		params.Add("q", "timestamp>"+d.lastTime.Format(ccTimeStampFormat))
	}
	err := fetch(params)
	if err != nil {
		return err
	}

	for nextURL != nil {
		t, err := url.Parse(*nextURL)
		if err != nil {
			return fmt.Errorf("failed to parse params out of %s: %s", nextURL, err)
		}

		err = fetch(t.Query())
		if err != nil {
			return err
		}
	}

	if len(resources) == 0 {
		d.saveLatestTime(time.Now())
	}

	for _, resource := range resources {
		eventType, ok := eventTypes[resource.Entity.Type]
		if !ok {
			continue
		}

		t, err := time.Parse(time.RFC3339, resource.Entity.Timestamp)
		if err != nil {
			return err
		}

		d.saveLatestTime(t)

		cmd := CreateHandleResponseCommand(resource , isTimestampSet , eventType , t, d.version, d.UUID, d.CcHost, d.httpClient, d.analyticsClient)
		err = cmd.HandleResponse()
		if err != nil {
			return err
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

type ResponseCommand interface {
	HandleResponse() error
}

func CreateHandleResponseCommand(resource cmd.Resource, isTimestampSet bool, eventType string, t time.Time, version string, uuid string, ccHost string, httpClient *http.Client,
analyticsClient analytics.Client) ResponseCommand {
	switch eventType {
	case "app created":
		return  &cmd.AppCreatedCmd {
			Resource: resource,
			IsTimestampSet: isTimestampSet,
			Version: version,
			Uuid: uuid,
			EventType: eventType,
			T: t,
			CcHost: ccHost,
			HttpClient: httpClient,
			AnalyticsClient:analyticsClient,
		}
	case "service created":
		return &cmd.ServiceCreatedCmd {
			Resource: resource,
			IsTimestampSet: isTimestampSet,
			Version: version,
			Uuid: uuid,
			EventType: eventType,
			T: t,
			CcHost: ccHost,
			HttpClient: httpClient,
			AnalyticsClient:analyticsClient,
		}
	}

	return nil
}
