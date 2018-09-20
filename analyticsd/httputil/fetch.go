package httputil

import (
	"encoding/json"
	"fmt"
	"gopkg.in/segmentio/analytics-go.v3"
	"io/ioutil"
	"net/http"
	"net/url"
	"runtime"
	"time"
)

func Fetch(ccHost string, apiEndpoint string, version string, uuid string, params url.Values, httpClient *http.Client, analyticsClient analytics.Client, dest interface{}) error {
	req, err := http.NewRequest(http.MethodGet, ccHost+apiEndpoint, nil)
	if err != nil {
		return err
	}

	req.URL.RawQuery = params.Encode()

	resp, err := httpClient.Do(req)
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
			"version": version,
		}

		err := analyticsClient.Enqueue(analytics.Track{
			UserId:     uuid,
			Event:      "analytics error",
			Timestamp:  time.Now().UTC(),
			Properties: properties,
		})

		if err != nil {
			return fmt.Errorf("failed to send analytics: %v", err)
		}

		//think about logging error anyway if failed to contact cc
		//instead of return nil
		return nil
	}

	return json.Unmarshal(contents, dest)
}
