package cmd_test

import (
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

var _ = Describe("HandleAppCrash", func() {

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

	It("should send service_bind", func() {
		var mockServiceInstancesResponse = `
			{
				"entity": {
					"service_url: "/v2/services/e967701c-d82f-49bc-8b56-dc70938f3948"
				}
			}
			`
		var mockServicesResponse = `
			{
				"entity": {
					"label": "p-mysql",
				}
			}
			`

		t := time.Date(2018, 8, 9, 8, 8, 8, 0, time.UTC)

		ccServer.AppendHandlers(
			ghttp.CombineHandlers(
				ghttp.VerifyRequest(http.MethodGet, "/v2/service_instances/some-service-instance-guid"),
				ghttp.RespondWith(http.StatusOK, mockServiceInstancesResponse),
			),
			ghttp.CombineHandlers(
				ghttp.VerifyRequest(http.MethodGet, "/v2/services/e967701c-d82f-49bc-8b56-dc70938f3948"),
				ghttp.RespondWith(http.StatusOK, mockServicesResponse),
			),
		)

		mockAnalytics.EXPECT().Enqueue(analytics.Track{
			UserId:    "some-user-uuid",
			Event:     "app bound to service",
			Timestamp: t,
			Properties: map[string]interface{}{
				"service": "p-mysql",
				"os":      runtime.GOOS,
				"version": "some-version",
			},
		})

		mockResource := cmd.Resource{
			Entity: cmd.Entity{
				Metadata: cmd.Metadata{
					Request: cmd.Request{
						Relationships: cmd.Relationships{
							ServiceInstance: cmd.ServiceInstance{
								Data: cmd.Data{
									Guid: "some-service-instance-guid",
								},
							},
						},
					},
				},
			},
		}

		command := cmd.CreateHandleResponseCommand(mockResource, true, "app bound to service", t, "some-version", "some-user-uuid", ccServer.URL(), httpClient, mockAnalytics)
		command.HandleResponse()
	})
})

