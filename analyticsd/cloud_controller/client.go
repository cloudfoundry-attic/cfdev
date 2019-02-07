package cloud_controller

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const ccTimeStampFormat = "2006-01-02T15:04:05Z"

//go:generate mockgen -package mocks -destination mocks/analytics.go gopkg.in/segmentio/analytics-go.v3 Client

type Client struct {
	host            string
	logger          *log.Logger
	httpClient      *http.Client
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
	"audit.organization.create",
	"audit.space.create",
	"audit.service_instance.create",
	"audit.service_binding.create",
	"audit.service_broker.create",
	"audit.user_provided_service_instance.create",
	"audit.route.create",
}

func New(host string, logger *log.Logger, httpClient *http.Client) *Client {
	return &Client{
		host:            host,
		logger:          logger,
		httpClient:      httpClient,
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

	c.logger.Printf("failed to contact cc api: [%v] %s", resp.Status, contents)
	return nil
}
