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

var _ = Describe("HandleAppCreated", func() {
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

	Context("when there are subsequent events", func() {
		It("sends the subsequent events", func() {
			t := time.Date(2018, 8, 9, 8, 8, 8, 0, time.UTC)

			mockAnalytics.EXPECT().Enqueue(analytics.Track{
				UserId:    "some-user-uuid",
				Event:     "app created",
				Timestamp: t,
				Properties: map[string]interface{}{
					"buildpack": "ruby",
					"os":        runtime.GOOS,
					"version":   "some-version",
				},
			})

			mockResource := cmd.Resource{
				Entity: cmd.Entity{
					Metadata: cmd.Metadata{
						Request: cmd.Request{
							Buildpack: "ruby_buildpack",
						},
					},
				},
			}

			cmd := daemon.CreateHandleResponseCommand(mockResource, true, "app created", t, "some-version", "some-user-uuid", ccServer.URL(), httpClient, mockAnalytics)
			cmd.HandleResponse()
		})
	})
})

