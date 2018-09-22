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

type Daemon2 struct {
	UUID            string
	version         string
	ccClient        *cloud_controller.Client
	analyticsClient analytics.Client
	pollingInterval time.Duration
	logger          *log.Logger
	lastTime        time.Time
	doneChan        chan bool
}

func New2(
	ccHost string,
	UUID string,
	version string,
	writer io.Writer,
	httpClient *http.Client,
	analyticsClient analytics.Client,
	pollingInterval time.Duration,
) *Daemon2 {
	logger := log.New(writer, "[ANALYTICSD] ", log.LstdFlags)
	ccClient := cloud_controller.New(ccHost, logger, httpClient, analyticsClient, UUID, version)

	return &Daemon2{
		UUID:            UUID,
		version:         version,
		ccClient:        ccClient,
		analyticsClient: analyticsClient,
		pollingInterval: pollingInterval,
		logger:          logger,
		doneChan:        make(chan bool, 1),
	}
}

func (d *Daemon2) Start() {
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

func (d *Daemon2) Stop() {
	d.doneChan <- true
}

func (d *Daemon2) do() error {
	events, err := d.ccClient.FetchEvents(d.lastTime)
	if err != nil {
		return err
	}

	for _, event := range events {
		d.saveLatestTime(event.Timestamp)

		cmd, exists := command.New(event.Type, d.ccClient, d.analyticsClient, d.lastTime, d.UUID, d.version, d.logger)
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

func (d *Daemon2) saveLatestTime(t time.Time) {
	if t.After(d.lastTime) {
		d.lastTime = t
	}
}
