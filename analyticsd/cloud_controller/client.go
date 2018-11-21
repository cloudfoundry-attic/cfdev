package cloud_controller

import (
	"encoding/json"
	"fmt"
	"gopkg.in/segmentio/analytics-go.v3"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"runtime"
	"strings"
	"time"
)

const ccTimeStampFormat = "2006-01-02T15:04:05Z"

//go:generate mockgen -package mocks -destination mocks/analytics.go gopkg.in/segmentio/analytics-go.v3 Client

type Client struct {
	host            string
	logger          *log.Logger
	httpClient      *http.Client
	analyticsClient analytics.Client
	userUUID        string
	version         string
}

type Event struct {
	Type      string
	Timestamp time.Time
	Metadata  json.RawMessage
}

type eventResponse struct {
	NextURL   *string `json:"next_url"`
	Resources []struct {
		Entity struct {
			Type      string
			Timestamp string
			Metadata  json.RawMessage
		}
	}
}

var eventTypes = []string{
	"audit.app.create",
	"audit.app.restage",
	"app.crash",
	"audit.service_instance.create",
	"audit.service_binding.create",
	"audit.service_broker.create",
	"audit.user_provided_service_instance.create",
	"audit.route.create",
}

func New(host string, logger *log.Logger, httpClient *http.Client, analyticsClient analytics.Client, userUUID string, version string) *Client {
	return &Client{
		host:            host,
		logger:          logger,
		httpClient:      httpClient,
		analyticsClient: analyticsClient,
		userUUID:        userUUID,
		version:         version,
	}
}

func (c *Client) FetchLatestTime() time.Time {
	params := url.Values{}
	params.Add("order-by", "timestamp")
	params.Add("order-direction", "desc")
	// consider limiting the results with 'results-per-page'

	var result struct {
		Resources []struct {
			Entity struct {
				Timestamp string
			}
		}
	}

	c.logger.Println("Fetching latest timestamp from Cloud Controller...")

	err := c.Fetch("/v2/events", params, &result)
	if err != nil {
		return time.Now().UTC()
	}

	if len(result.Resources) == 0 {
		return time.Now().UTC()
	}

	t, _ := time.Parse(time.RFC3339, result.Resources[0].Entity.Timestamp)
	c.logger.Printf("Using timestamp of %v to mark new events\n", t)
	return t
}

func (c *Client) FetchEvents(timeStamp time.Time) ([]Event, error) {
	var (
		events  []Event
		nextURL *string = nil
		fetch           = func(params url.Values) error {
			var response eventResponse
			err := c.Fetch("/v2/events", params, &response)
			if err != nil {
				return err
			}

			for _, resource := range response.Resources {
				t, _ := time.Parse(time.RFC3339, resource.Entity.Timestamp)

				events = append(events, Event{
					Type:      resource.Entity.Type,
					Timestamp: t,
					Metadata:  resource.Entity.Metadata,
				})
			}

			nextURL = response.NextURL
			return nil
		}
	)

	params := url.Values{}
	params.Add("q", "type IN "+strings.Join(eventTypes, ","))
	params.Add("q", "timestamp>"+timeStamp.Format(ccTimeStampFormat))

	err := fetch(params)
	if err != nil {
		return nil, err
	}

	for nextURL != nil {
		t, err := url.Parse(*nextURL)
		if err != nil {
			return nil, err
		}

		err = fetch(t.Query())
		if err != nil {
			return nil, err
		}
	}

	return events, nil
}

func (c *Client) Fetch(path string, params url.Values, dest interface{}) error {
	url := c.host + path

	c.logger.Printf("Making request to %q with params: %v...\n", url, params)

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return err
	}

	req.URL.RawQuery = params.Encode()

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to query cloud controller: %s", err)
	}

	contents, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	c.logger.Printf("Received status code [%s] from url: %q\n", resp.Status, url)

	if resp.StatusCode == http.StatusOK {
		return json.Unmarshal(contents, dest)
	}

	var properties = analytics.Properties{
		"message": fmt.Sprintf("failed to contact cc api: [%v] %s", resp.Status, contents),
		"os":      runtime.GOOS,
		"version": c.version,
	}

	c.logger.Println("Sending an error to segment.io...")

	// Still not sure if sending every error to segment
	// is preferred behavior
	err = c.analyticsClient.Enqueue(analytics.Track{
		UserId:     c.userUUID,
		Event:      "analytics error",
		Timestamp:  time.Now().UTC(),
		Properties: properties,
	})

	if err != nil {
		c.logger.Printf("Failed to send analytics error: %v\n", err)
	}

	return nil
}
