package command_test

import (
	"code.cloudfoundry.org/cfdev/analyticsd/command"
	"code.cloudfoundry.org/cfdev/analyticsd/command/mocks"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	"gopkg.in/segmentio/analytics-go.v3"
	"io/ioutil"
	"log"
	"runtime"
	"time"
)

var _ = Describe("AppCrash", func() {
	var (
		cmd            *command.AppCrash
		mockController *gomock.Controller
		mockAnalytics  *mocks.MockClient
	)

	BeforeEach(func() {
		mockController = gomock.NewController(GinkgoT())
		mockAnalytics = mocks.NewMockClient(mockController)

		cmd = &command.AppCrash{
			Logger:          log.New(ioutil.Discard, "", log.LstdFlags),
			AnalyticsClient: mockAnalytics,
			TimeStamp:       time.Date(2018, 8, 8, 8, 8, 8, 0, time.UTC),
			UUID:            "some-user-uuid",
			Version:         "some-version",
			OSVersion:       "some-os-version",
		}
	})

	AfterEach(func() {
		mockController.Finish()
	})

	Context("when an app crash event occurs", func() {
		It("sends the the crash event to segment.io", func() {
			mockAnalytics.EXPECT().Enqueue(analytics.Track{
				UserId:    "some-user-uuid",
				Event:     "app push failed",
				Timestamp: time.Date(2018, 8, 8, 8, 8, 8, 0, time.UTC),
				Properties: map[string]interface{}{
					"os":             runtime.GOOS,
					"plugin_version": "some-version",
					"os_version":     "some-os-version",
				},
			})

			body := []byte(`
			{
                  "instance": "d73d6816-3101-4efb-4a9b-c4c1",
                  "index": 0,
                  "cell_id": "f5002113-527c-4bf9-b5e4-867f68a4ecfc",
                  "exit_description": "APP/PROC/WEB: Exited with status 1",
                  "reason": "CRASHED"
            }`)

			cmd.HandleResponse(body)
		})
	})
})
