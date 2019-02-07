package daemon

import (
	"code.cloudfoundry.org/cfdev/analyticsd/cloud_controller"
	"code.cloudfoundry.org/cfdev/analyticsd/command"
	"gopkg.in/segmentio/analytics-go.v3"
	"io"
	"log"
	"net/http"
	"time"
)

//go:generate mockgen -package mocks -destination mocks/analytics.go gopkg.in/segmentio/analytics-go.v3 Client

type Daemon struct {
	UUID            string
	pluginVersion   string
	osVersion       string
	isBehindProxy   string
	ccClient        *cloud_controller.Client
	analyticsClient analytics.Client
	pollingInterval time.Duration
	logger          *log.Logger
	lastTime        time.Time
	doneChan        chan bool
}

func New(
	ccHost string,
	UUID string,
	pluginVersion string,
	osVersion string,
	writer io.Writer,
	httpClient *http.Client,
	analyticsClient analytics.Client,
	pollingInterval time.Duration,
) *Daemon {
	logger := log.New(writer, "[ANALYTICSD] ", log.LstdFlags)
	ccClient := cloud_controller.New(ccHost, logger, httpClient)

	return &Daemon{
		UUID:            UUID,
		pluginVersion:   pluginVersion,
		osVersion:       osVersion,
		ccClient:        ccClient,
		analyticsClient: analyticsClient,
		pollingInterval: pollingInterval,
		logger:          logger,
		doneChan:        make(chan bool, 1),
	}
}

func (d *Daemon) Start() {
	t := d.ccClient.FetchLatestTime()
	d.saveLatestTime(t)

	ticker := time.NewTicker(d.pollingInterval)

	for {
		select {
		case <-d.doneChan:
			return
		case <-ticker.C:
			err := d.do()
			if err != nil {
				d.logger.Println(err)
			}
		}
	}
}

func (d *Daemon) Stop() {
	err := d.do()
	if err != nil {
		d.logger.Println(err)
	}
	d.doneChan <- true
}

func (d *Daemon) do() error {
	events, err := d.ccClient.FetchEvents(d.lastTime)
	if err != nil {
		return err
	}

	for _, event := range events {
		d.saveLatestTime(event.Timestamp)

		cmd, exists := command.New(event.Type, d.ccClient, d.analyticsClient, event.Timestamp, d.UUID, d.pluginVersion, d.osVersion, d.logger)
		if !exists {
			continue
		}

		err = cmd.HandleResponse(event.Metadata)
		if err != nil {
			return err
		}
	}

	return nil
}

func (d *Daemon) saveLatestTime(t time.Time) {
	if t.After(d.lastTime) {
		d.lastTime = t
	}
}
