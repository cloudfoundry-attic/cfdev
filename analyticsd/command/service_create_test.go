package command_test

import (
	"code.cloudfoundry.org/cfdev/analyticsd/command"
	commandMocks "code.cloudfoundry.org/cfdev/analyticsd/command/mocks"
	"code.cloudfoundry.org/cfdev/analyticsd/segment"
	"code.cloudfoundry.org/cfdev/analyticsd/segment/mocks"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"gopkg.in/segmentio/analytics-go.v3"
	"io/ioutil"
	"log"
	"time"
)

var _ = Describe("ServiceCreate", func() {
	var (
		cmd            *command.ServiceCreate
		mockController *gomock.Controller
		mockAnalytics  *mocks.MockClient
		mockCCClient   *commandMocks.MockCloudControllerClient
	)

	BeforeEach(func() {
		mockController = gomock.NewController(GinkgoT())
		mockAnalytics = mocks.NewMockClient(mockController)
		mockCCClient = commandMocks.NewMockCloudControllerClient(mockController)

		segmentClient := segment.New(
			mockAnalytics,
			"some-user-uuid",
			"some-version",
			"some-os-version",
			time.Date(2018, 8, 8, 8, 8, 8, 0, time.UTC),
		)

		cmd = &command.ServiceCreate{
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
			MatchFetch(mockCCClient, "/v2/service_plans/some-service-plan-guid", `
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
				Expect(event.Event).To(Equal("created service"))
				Expect(event.Timestamp).To(Equal(time.Date(2018, 8, 8, 8, 8, 8, 0, time.UTC)))
				Expect(event.Properties).To(HaveKeyWithValue("service", "mysql"))
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

	Context("when the service instance is NOT whitelisted", func() {
		It("does not send the service information to NOT segment.io", func() {
			MatchFetch(mockCCClient, "/v2/service_plans/some-service-plan-guid", `
				{
            		"entity": {
						"service_url": "/v2/some_service_url"
                    }
				}
				`)

			MatchFetch(mockCCClient, "/v2/some_service_url", `
				{
            		"entity": {
						"label": "my-special-sql"
                    }
				}
				`)

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
