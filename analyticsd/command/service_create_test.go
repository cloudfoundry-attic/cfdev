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

var _ = Describe("ServiceCreate", func() {
	var (
		cmd            *command.ServiceCreate
		mockController *gomock.Controller
		mockAnalytics  *mocks.MockClient
		mockCCclient   *mocks.MockCloudControllerClient
	)

	BeforeEach(func() {
		mockController = gomock.NewController(GinkgoT())
		mockAnalytics = mocks.NewMockClient(mockController)
		mockCCclient = mocks.NewMockCloudControllerClient(mockController)

		cmd = &command.ServiceCreate{
			Logger: log.New(ioutil.Discard, "", log.LstdFlags),
			CCclient: mockCCclient,
			AnalyticsClient: mockAnalytics,
			TimeStamp: time.Date(2018, 8, 8, 8, 8, 8, 0, time.UTC),
			UUID: "some-user-uuid",
			Version: "some-version",
		}
	})

	AfterEach(func() {
		mockController.Finish()
	})

	Context("when the service instance is whitelisted", func() {
		It("sends the service information to segment.io", func() {
			MatchFetch(mockCCclient, "/v2/service_plans/some-service-plan-guid", `
				{
            		"entity": {
						"service_url": "/v2/some_service_url"
                    }
				}
				`)

			MatchFetch(mockCCclient, "/v2/some_service_url", `
				{
            		"entity": {
						"label": "mysql"
                    }
				}
				`)

			mockAnalytics.EXPECT().Enqueue(analytics.Track{
				UserId:    "some-user-uuid",
				Event:     "app bound to service",
				Timestamp: time.Date(2018, 8, 8, 8, 8, 8, 0, time.UTC),
				Properties: map[string]interface{}{
					"service":   "mysql",
					"os":        runtime.GOOS,
					"version":   "some-version",
				},
			})

			body := []byte(`
			{
				"request": {
					"service_plan_guid": "some-service-plan-guid" 
				}
			}`)

			cmd.HandleResponse(body)
		})
	})
})