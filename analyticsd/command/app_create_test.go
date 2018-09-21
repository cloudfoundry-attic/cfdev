package command_test

import (
	"code.cloudfoundry.org/cfdev/analyticsd/command/mocks"
	"code.cloudfoundry.org/cfdev/analyticsd/command"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	"gopkg.in/segmentio/analytics-go.v3"
	"io/ioutil"
	"log"
	"runtime"
	"time"
)

var _ = Describe("AppCreate", func() {
	var (
		cmd            *command.AppCreate
		mockController *gomock.Controller
		mockAnalytics  *mocks.MockClient
	)

	BeforeEach(func() {
		mockController = gomock.NewController(GinkgoT())
		mockAnalytics = mocks.NewMockClient(mockController)

		cmd = &command.AppCreate{
			Logger: log.New(ioutil.Discard, "", log.LstdFlags),
			AnalyticsClient: mockAnalytics,
			TimeStamp: time.Date(2018, 8, 8, 8, 8, 8, 0, time.UTC),
			UUID: "some-user-uuid",
			Version: "some-version",
		}
	})

	AfterEach(func() {
		mockController.Finish()
	})

	Context("when the buildpack is whitelisted", func() {
		It("sends the buildpack information to segment.io", func() {
			mockAnalytics.EXPECT().Enqueue(analytics.Track{
				UserId:    "some-user-uuid",
				Event:     "app created",
				Timestamp: time.Date(2018, 8, 8, 8, 8, 8, 0, time.UTC),
				Properties: map[string]interface{}{
					"buildpack": "go",
					"os":        runtime.GOOS,
					"version":   "some-version",
				},
			})

			body := []byte(`
			{
				"request": {
					"buildpack": "go_buildpack"
				}
			}`)

			cmd.HandleResponse(body)
		})
	})

	Context("when the buildpack is not whitelisted", func() {
		It("sends the buildpack information to segment.io", func() {
			mockAnalytics.EXPECT().Enqueue(analytics.Track{
				UserId:    "some-user-uuid",
				Event:     "app created",
				Timestamp: time.Date(2018, 8, 8, 8, 8, 8, 0, time.UTC),
				Properties: map[string]interface{}{
					"buildpack": "custom",
					"os":        runtime.GOOS,
					"version":   "some-version",
				},
			})

			body := []byte(`
			{
				"request": {
					"buildpack": "some-unexpected-buildpack"
				}
			}`)

			cmd.HandleResponse(body)
		})
	})
})