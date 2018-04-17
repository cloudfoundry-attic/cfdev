package cfanalytics_test

import (
	"runtime"

	"code.cloudfoundry.org/cfdev/cfanalytics"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"gopkg.in/segmentio/analytics-go.v3"
)

type MockClient struct {
	WasCalledWith analytics.Message
}

func (mc *MockClient) Enqueue(message analytics.Message) error {
	mc.WasCalledWith = message
	return nil
}

func (mc *MockClient) Close() error {
	return nil
}

func (mc *MockClient) SendAnalytics() error {
	return nil
}

var _ = Describe("Startevent", func() {
	It("sends data to client", func() {
		mockClient := MockClient{
			WasCalledWith: analytics.Track{},
		}

		expectProperties := analytics.NewProperties()
		expectProperties.Set("os", runtime.GOOS)
		expectProperties.Set("version", "2.0")
		expectProperties.Set("type", "cf")

		expected := analytics.Track{
			UserId:     "my user id",
			Event:      "start",
			Properties: expectProperties,
		}

		analyticsCmd := cfanalytics.AnalyticsEvent{
			SegClient: &mockClient,
			Event:     "start",
			UserId:    "my user id",
			Type:      "cf",
			OS:        runtime.GOOS,
			Version:   "2.0",
		}

		analyticsCmd.SendAnalytics()

		Expect(mockClient.WasCalledWith).To(Equal(expected))
	})
})
