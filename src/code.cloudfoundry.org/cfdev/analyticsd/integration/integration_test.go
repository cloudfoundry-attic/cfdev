package integration

import (
	"code.cloudfoundry.org/cfdev/analyticsd/daemon"
	"code.cloudfoundry.org/cfdev/analyticsd/daemon/mocks"
	"errors"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/ghttp"
	"gopkg.in/segmentio/analytics-go.v3"
	"net/http"
	"runtime"
	"time"
)

var _ = Describe("Integration", func() {

	var (
		aDaemon        *daemon.Daemon
		ccServer       *ghttp.Server
		mockController *gomock.Controller
		mockAnalytics  *mocks.MockClient
		buffer         *gbytes.Buffer
		httpClient     *http.Client
	)

	BeforeEach(func() {
		ccServer = ghttp.NewServer()

		mockController = gomock.NewController(GinkgoT())
		mockAnalytics = mocks.NewMockClient(mockController)
		buffer = gbytes.NewBuffer()
		httpClient = &http.Client{}

		t, _ := time.Parse(time.RFC3339, "2017-08-08T08:08:08Z")
		aDaemon = daemon.New(
			ccServer.URL(),
			"some-user-uuid",
			buffer,
			httpClient,
			mockAnalytics,
			time.Second,
			t,
		)
	})

	AfterEach(func() {
		aDaemon.Stop()
		ccServer.Close()

		mockController.Finish()
	})

	startDaemon := func() {
		go func() {
			defer GinkgoRecover()

			aDaemon.Start()
		}()
	}

	Describe("app pushes", func() {
		Context("when an app push event has occurred", func() {

			BeforeEach(func() {
				ccServer.AppendHandlers(ghttp.CombineHandlers(
					ghttp.VerifyRequest(http.MethodGet, "/v2/events", "q=type%20IN%20audit.app.create&q=timestamp>2017-08-08T08:08:08Z"),
					ghttp.RespondWith(http.StatusOK, fixturePushApp),
				))
			})

			It("it relays the event to segment.io", func() {
				mockAnalytics.EXPECT().Enqueue(analytics.Track{
					UserId:    "some-user-uuid",
					Event:     "app push",
					Timestamp: time.Date(2018, 8, 8, 8, 8, 8, 0, time.UTC),
					Properties: map[string]interface{}{
						"buildpack": "ruby_buildpack",
						"os":        runtime.GOOS,
					},
				})

				mockAnalytics.EXPECT().Enqueue(analytics.Track{
					UserId:    "some-user-uuid",
					Event:     "app push",
					Timestamp: time.Date(2018, 9, 9, 9, 9, 9, 0, time.UTC),
					Properties: map[string]interface{}{
						"buildpack": "go_buildpack",
						"os":        runtime.GOOS,
					},
				})

				startDaemon()
				<-time.After(1500 * time.Millisecond)
			})
		})
	})

	Describe("successive metrics", func() {
		BeforeEach(func() {
			ccServer.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest(http.MethodGet, "/v2/events","q=type%20IN%20audit.app.create&q=timestamp>2017-08-08T08:08:08Z"),
					ghttp.RespondWith(http.StatusOK, fixtureSequentialResponse1),
				),
				ghttp.CombineHandlers(
					ghttp.VerifyRequest(http.MethodGet, "/v2/events", "q=type%20IN%20audit.app.create&q=timestamp>2018-08-08T08:08:08Z"),
					ghttp.RespondWith(http.StatusOK, fixtureSequentialResponse2),
				),
			)
		})

		It("are filtered out because of constantly retrieving newer records by one second", func() {
			mockAnalytics.EXPECT().Enqueue(gomock.Any()).Times(2)

			startDaemon()
			<-time.After(2500 * time.Millisecond)
		})
	})

	Describe("no last time set, successive metrics sent", func() {
		BeforeEach(func() {
			aDaemon = daemon.New(
				ccServer.URL(),
				"some-user-uuid",
				buffer,
				httpClient,
				mockAnalytics,
				time.Second,
				time.Time{},
			)

			ccServer.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest(http.MethodGet, "/v2/events","q=type%20IN%20audit.app.create"),
					ghttp.RespondWith(http.StatusOK, fixtureSequentialResponse1),
				),
				ghttp.CombineHandlers(
					ghttp.VerifyRequest(http.MethodGet, "/v2/events", "q=type%20IN%20audit.app.create&q=timestamp>2018-08-08T08:08:08Z"),
					ghttp.RespondWith(http.StatusOK, fixtureSequentialResponse2),
				),
			)
		})

		It("are filtered out because of constantly retrieving newer records by one second", func() {
			mockAnalytics.EXPECT().Enqueue(gomock.Any()).Times(1)

			startDaemon()
			<-time.After(2500 * time.Millisecond)
		})
	})

	Describe("requests to Cloud Controller gives a non successful status code", func() {
		BeforeEach(func() {
			ccServer.AppendHandlers(ghttp.CombineHandlers(
				ghttp.VerifyRequest(http.MethodGet, "/v2/events"),
				ghttp.RespondWith(http.StatusUnauthorized, `some non-authorized error`),
			))
		})

		It("it relays the error to segment.io", func() {
			mockAnalytics.EXPECT().Enqueue(gomock.Any()).Do(func(msg analytics.Track) {
				Expect(msg.UserId).To(Equal("some-user-uuid"))
				Expect(msg.Event).To(Equal("analytics error"))
				Expect(msg.Properties["message"]).To(ContainSubstring("failed to contact cc api: [401 Unauthorized] some non-authorized error"))
			}).Times(1)

			startDaemon()
			<-time.After(1500 * time.Millisecond)
		})

		It("it logs the error", func() {
			mockAnalytics.EXPECT().Enqueue(gomock.Any()).AnyTimes().Return(errors.New("some api error"))

			startDaemon()
			<-time.After(1500 * time.Millisecond)
			Expect(buffer).To(gbytes.Say(`failed to send analytics: some api error`))
		})

	})

	Describe("when making a request to Cloud Controller fails", func() {
		BeforeEach(func() {
			httpClient.Timeout = time.Millisecond
			ccServer.AppendHandlers(func(w http.ResponseWriter, req *http.Request) {
				time.Sleep(time.Second)
			})
		})

		It("it logs the error", func() {
			mockAnalytics.EXPECT().Enqueue(gomock.Any()).Times(0)

			startDaemon()
			<-time.After(1500 * time.Millisecond)
			Expect(buffer).To(gbytes.Say(`failed to query cloud controller:`))
		})

	})

	Describe("when sending the metrics fails", func() {
		BeforeEach(func() {
			ccServer.AppendHandlers(ghttp.CombineHandlers(
				ghttp.VerifyRequest(http.MethodGet, "/v2/events"),
				ghttp.RespondWith(http.StatusOK, fixturePushApp),
			))
		})

		It("it logs the error", func() {
			mockAnalytics.EXPECT().Enqueue(gomock.Any()).AnyTimes().Return(errors.New("some api error"))

			startDaemon()
			<-time.After(1500 * time.Millisecond)
			Expect(buffer).To(gbytes.Say(`failed to send analytics: some api error`))
		})

	})

	Describe("unexpected event types", func() {
		BeforeEach(func() {
			ccServer.AppendHandlers(ghttp.CombineHandlers(
				ghttp.VerifyRequest(http.MethodGet, "/v2/events"),
				ghttp.RespondWith(http.StatusOK, fixtureUnexpected),
			))
		})

		It("does not send any metrics to segment.io", func() {
			mockAnalytics.EXPECT().Enqueue(gomock.Any()).Times(0)

			startDaemon()
			<-time.After(1500 * time.Millisecond)
		})
	})
})
