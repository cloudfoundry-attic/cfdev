package command_test

import (
	"code.cloudfoundry.org/cfdev/pkg/analyticsd/command"
	commmandMocks "code.cloudfoundry.org/cfdev/pkg/analyticsd/command/mocks"
	"code.cloudfoundry.org/cfdev/pkg/analyticsd/segment"
	"code.cloudfoundry.org/cfdev/pkg/analyticsd/segment/mocks"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"gopkg.in/segmentio/analytics-go.v3"
	"io/ioutil"
	"log"
	"time"
)

var _ = Describe("ServiceBind", func() {
	var (
		cmd            *command.ServiceBind
		mockController *gomock.Controller
		mockAnalytics  *mocks.MockClient
		mockCCClient   *commmandMocks.MockCloudControllerClient
	)

	BeforeEach(func() {
		mockController = gomock.NewController(GinkgoT())
		mockAnalytics = mocks.NewMockClient(mockController)
		mockCCClient = commmandMocks.NewMockCloudControllerClient(mockController)
		segmentClient := segment.New(
			mockAnalytics,
			"some-user-uuid",
			"some-version",
			"some-os-version",
			time.Date(2018, 8, 8, 8, 8, 8, 0, time.UTC),
		)

		cmd = &command.ServiceBind{
			Logger:          log.New(ioutil.Discard, "", log.LstdFlags),
			CCClient:        mockCCClient,
			AnalyticsClient: segmentClient,
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

			mockAnalytics.EXPECT().Enqueue(gomock.Any()).Do(func(event analytics.Track) {
				Expect(event.UserId).To(Equal("some-user-uuid"))
				Expect(event.Event).To(Equal("app bound to service"))
				Expect(event.Timestamp).To(Equal(time.Date(2018, 8, 8, 8, 8, 8, 0, time.UTC)))
				Expect(event.Properties).To(HaveKeyWithValue("service", "mysql"))
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
