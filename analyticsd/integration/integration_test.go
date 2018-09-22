package integration

import (
	"fmt"
	"net/http"
	"net/url"
	"runtime"
	"strings"
	"time"

	"code.cloudfoundry.org/cfdev/analyticsd/daemon"
	"code.cloudfoundry.org/cfdev/analyticsd/daemon/mocks"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/ghttp"
	"gopkg.in/segmentio/analytics-go.v3"
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

		aDaemon = daemon.New(
			ccServer.URL(),
			"some-user-uuid",
			"some-version",
			buffer,
			httpClient,
			mockAnalytics,
			time.Second,
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
		Context("when there are historical events (events found on first request)", func() {
			BeforeEach(func() {
				ccServer.AppendHandlers(ghttp.CombineHandlers(
					ghttp.VerifyRequest(http.MethodGet, "/v2/events"),

					ghttp.RespondWith(http.StatusOK, fakeResponse([]string{
						fakePushEvent("2018-08-08T08:08:08Z", "some-buildpack"),
						fakePushEvent("2018-08-08T08:08:07Z", "some-other-buildpack"),
					})),
				))
			})

			It("does not send those events", func() {
				startDaemon()
				<-time.After(500 * time.Millisecond)
			})

			Context("when there are subsequent events", func() {
				BeforeEach(func() {
					ccServer.AppendHandlers(
						ghttp.CombineHandlers(
							ghttp.VerifyRequest(http.MethodGet, "/v2/events"),
							func(w http.ResponseWriter, req *http.Request) {
								rawQuery := req.URL.RawQuery
								Expect(rawQuery).To(ContainSubstring("audit.app.create"))
								Expect(rawQuery).To(ContainSubstring("audit.service_instance.create"))
								Expect(rawQuery).To(ContainSubstring("timestamp%3E2018-08-08T08%3A08%3A08Z"))
							},
							ghttp.RespondWith(http.StatusOK, fakeResponse([]string{
								fakePushEvent("2018-08-09T08:08:08Z", "ruby_buildpack"),
								fakePushEvent("2018-08-08T09:07:08Z", "go_buildpack"),
							})),
						),
						ghttp.CombineHandlers(
							ghttp.VerifyRequest(http.MethodGet, "/v2/events"),
							func(w http.ResponseWriter, req *http.Request) {
								rawQuery := req.URL.RawQuery
								Expect(rawQuery).To(ContainSubstring("audit.app.create"))
								Expect(rawQuery).To(ContainSubstring("audit.service_instance.create"))
								Expect(rawQuery).To(ContainSubstring("timestamp%3E2018-08-09T08%3A08%3A08Z"))
							},
							ghttp.RespondWith(http.StatusOK, fakeResponse([]string{
								fakePushEvent("2018-08-10T08:08:08Z", "java_buildpack"),
								fakePushEvent("2018-08-11T08:08:08Z", "nodejs_buildpack"),
							})),
						))
				})

				It("sends the events and continues polling", func() {
					mockAnalytics.EXPECT().Enqueue(analytics.Track{
						UserId:    "some-user-uuid",
						Event:     "app created",
						Timestamp: time.Date(2018, 8, 9, 8, 8, 8, 0, time.UTC),
						Properties: map[string]interface{}{
							"buildpack": "ruby",
							"os":        runtime.GOOS,
							"version":   "some-version",
						},
					})

					mockAnalytics.EXPECT().Enqueue(analytics.Track{
						UserId:    "some-user-uuid",
						Event:     "app created",
						Timestamp: time.Date(2018, 8, 8, 9, 7, 8, 0, time.UTC),
						Properties: map[string]interface{}{
							"buildpack": "go",
							"os":        runtime.GOOS,
							"version":   "some-version",
						},
					})

					mockAnalytics.EXPECT().Enqueue(analytics.Track{
						UserId:    "some-user-uuid",
						Event:     "app created",
						Timestamp: time.Date(2018, 8, 10, 8, 8, 8, 0, time.UTC),
						Properties: map[string]interface{}{
							"buildpack": "java",
							"os":        runtime.GOOS,
							"version":   "some-version",
						},
					})

					mockAnalytics.EXPECT().Enqueue(analytics.Track{
						UserId:    "some-user-uuid",
						Event:     "app created",
						Timestamp: time.Date(2018, 8, 11, 8, 8, 8, 0, time.UTC),
						Properties: map[string]interface{}{
							"buildpack": "nodejs",
							"os":        runtime.GOOS,
							"version":   "some-version",
						},
					})
					startDaemon()
					<-time.After(2030 * time.Millisecond)
				})
			})
		})

		Describe("when there are no historical events", func() {
			BeforeEach(func() {
				ccServer.AppendHandlers(ghttp.CombineHandlers(
					ghttp.VerifyRequest(http.MethodGet, "/v2/events"),

					ghttp.RespondWith(http.StatusOK, fakeResponse([]string{})),
				))
			})
			Context("when there are subsequent events", func() {
				BeforeEach(func() {
					ccServer.AppendHandlers(ghttp.CombineHandlers(
						ghttp.VerifyRequest(http.MethodGet, "/v2/events"),
						ghttp.RespondWith(http.StatusOK, fakeResponse([]string{
							fakePushEvent("2018-08-09T08:08:08Z", "ruby_buildpack"),
							fakePushEvent("2018-08-08T09:07:08Z", "go_buildpack"),
						})),
					))
				})
				It("sends the subsequent events", func() {
					mockAnalytics.EXPECT().Enqueue(analytics.Track{
						UserId:    "some-user-uuid",
						Event:     "app created",
						Timestamp: time.Date(2018, 8, 9, 8, 8, 8, 0, time.UTC),
						Properties: map[string]interface{}{
							"buildpack": "ruby",
							"os":        runtime.GOOS,
							"version":   "some-version",
						},
					})

					mockAnalytics.EXPECT().Enqueue(analytics.Track{
						UserId:    "some-user-uuid",
						Event:     "app created",
						Timestamp: time.Date(2018, 8, 8, 9, 7, 8, 0, time.UTC),
						Properties: map[string]interface{}{
							"buildpack": "go",
							"os":        runtime.GOOS,
							"version":   "some-version",
						},
					})

					startDaemon()
					<-time.After(1030 * time.Millisecond)
				})
			})

			Context("unexpected event types", func() {
				BeforeEach(func() {
					ccServer.AppendHandlers(ghttp.CombineHandlers(
						ghttp.VerifyRequest(http.MethodGet, "/v2/events"),
						ghttp.RespondWith(http.StatusOK, fakeResponse([]string{
							fakeEvent("unexpected.type", "some-timestamp", "some-buildpack"),
						})),
					))
				})

				It("does not send any metrics to segment.io", func() {
					mockAnalytics.EXPECT().Enqueue(gomock.Any()).Times(0)

					startDaemon()
					<-time.After(50 * time.Millisecond)
				})
			})
		})
	})

	Describe("HandleServiceCreated", func() {

	})
})

var pushAppEventTemplate = `
{
	"entity": {
		"type": "%s",
		"timestamp": "%s",
		"metadata": {
			"request": {
				"buildpack": "%s"
			}
		}
	}
}
`

func fakeEvent(eventType, timestamp, buildpack string) string {
	return fmt.Sprintf(pushAppEventTemplate, eventType, timestamp, buildpack)
}

func fakePushEvent(timestamp, buildpack string) string {
	return fakeEvent("audit.app.create", timestamp, buildpack)
}

var responseTemplate = `
{
    "next_url": %s,
	"resources": [%s]
}
`

func fakeResponse(events []string, args ...string) string {
	nextURL := "null"

	if len(args) > 0 {
		nextURL = fmt.Sprintf(`"%s"`, args[0])
	}

	return fmt.Sprintf(responseTemplate, nextURL, strings.Join(events, ","))
}

func urlContains(values url.Values, matches []string) {
	for _, match := range matches {
		Expect(values.Get("q")).To(ContainSubstring(match))
	}
}
