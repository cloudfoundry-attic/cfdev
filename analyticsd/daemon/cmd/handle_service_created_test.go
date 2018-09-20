package cmd_test

import (
	"code.cloudfoundry.org/cfdev/analyticsd/daemon"
	"code.cloudfoundry.org/cfdev/analyticsd/daemon/cmd"
	"code.cloudfoundry.org/cfdev/analyticsd/daemon/mocks"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	"github.com/onsi/gomega/ghttp"
	"gopkg.in/segmentio/analytics-go.v3"
	"net/http"
	"runtime"
	"time"
)

var _ = Describe("HandleServiceCreatedCommand", func() {

	var (
		mockController *gomock.Controller
		mockAnalytics  *mocks.MockClient
		ccServer       *ghttp.Server
		httpClient     *http.Client
	)

	BeforeEach(func() {
		ccServer = ghttp.NewServer()
		mockController = gomock.NewController(GinkgoT())
		mockAnalytics = mocks.NewMockClient(mockController)
		httpClient = &http.Client{}
	})

	It("Handles Service Created Event", func() {
		var mockResource = cmd.Resource{
			Entity: cmd.Entity{
				Metadata: cmd.Metadata{
					Request: cmd.Request{
						ServicePlanGUID: "myPlan",
					},
				},
			},
		}

		var mockServicePlanResponse = `
			{
				"entity": {
					"service_guid": "myServiceGuid"
				}
			}
 			`
		var mockServiceResponse = `
			{
				"entity": {
					"label": "myLabel"
				}
			}
			`

		t := time.Date(2018, 8, 9, 8, 8, 8, 0, time.UTC)

		ccServer.AppendHandlers(
			ghttp.CombineHandlers(
				ghttp.VerifyRequest(http.MethodGet, "/v2/service_plans/myPlan"),
				ghttp.RespondWith(http.StatusOK, mockServicePlanResponse),
			),
			ghttp.CombineHandlers(
				ghttp.VerifyRequest(http.MethodGet, "/v2/services/myServiceGuid"),
				ghttp.RespondWith(http.StatusOK, mockServiceResponse),
			),
		)

		mockAnalytics.EXPECT().Enqueue(analytics.Track{
			UserId:    "some-user-uuid",
			Event:     "service created",
			Timestamp: t,
			Properties: map[string]interface{}{
				"service": "myLabel",
				"os":      runtime.GOOS,
				"version": "some-version",
			},
		})

		cmd := daemon.CreateHandleResponseCommand(mockResource, true, "service created", t, "some-version", "some-user-uuid", ccServer.URL(), httpClient, mockAnalytics)
		cmd.HandleResponse()
	})
})
