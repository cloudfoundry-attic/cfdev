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

var _ = Describe("ServiceBind", func() {
	var (
		cmd            *command.ServiceBind
		mockController *gomock.Controller
		mockAnalytics  *mocks.MockClient
		mockCCClient   *mocks.MockCloudControllerClient
	)

	BeforeEach(func() {
		mockController = gomock.NewController(GinkgoT())
		mockAnalytics = mocks.NewMockClient(mockController)
		mockCCClient = mocks.NewMockCloudControllerClient(mockController)

		cmd = &command.ServiceBind{
			Logger:          log.New(ioutil.Discard, "", log.LstdFlags),
			CCClient:        mockCCClient,
			AnalyticsClient: mockAnalytics,
			TimeStamp:       time.Date(2018, 8, 8, 8, 8, 8, 0, time.UTC),
			UUID:            "some-user-uuid",
			Version:         "some-version",
		}
	})

	AfterEach(func() {
		mockController.Finish()
	})

	Context("when the service instance is whitelisted", func() {
		It("sends the service information to segment.io", func() {
			MatchFetch(mockCCClient, "/v2/service_instances/some-service-instance-guid", `
				{
            		"entity": {
						"service_url": "/v2/some_service_url"
                    }
				}
				`)

			MatchFetch(mockCCClient, "/v2/some_service_url", `
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
					"service": "mysql",
					"os":      runtime.GOOS,
					"version": "some-version",
				},
			})

			body := []byte(`
			{
				"request": {
					"relationships": {
						"service_instance": {
							"data": {
								"guid": "some-service-instance-guid"
							}
						}
                    }
				}
			}`)

			cmd.HandleResponse(body)
		})
	})

	Context("when the service instance is NOT whitelisted", func() {
		It("does not send the service information to segment.io", func() {
			MatchFetch(mockCCClient, "/v2/service_instances/some-service-instance-guid", `
				{
            		"entity": {
						"service_url": "/v2/some_service_url"
                    }
				}
				`)

			MatchFetch(mockCCClient, "/v2/some_service_url", `
				{
            		"entity": {
						"label": "non-white-listed-service"
                    }
				}
				`)

			body := []byte(`
			{
				"request": {
					"relationships": {
						"service_instance": {
							"data": {
								"guid": "some-service-instance-guid"
							}
						}
                    }
				}
			}`)

			cmd.HandleResponse(body)
		})
	})
})
