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
			"some-os-version",
			"false",
			buffer,
			httpClient,
			mockAnalytics,
			time.Second,
		)
	})

	AfterEach(func() {
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
						),
						ghttp.CombineHandlers(
							ghttp.VerifyRequest(http.MethodGet, "/v2/events"),
							ghttp.RespondWith(http.StatusOK, fakeResponse([]string{})),
						))
				})

				It("sends the events when analytics is stopped before polling interval completes", func() {
					mockAnalytics.EXPECT().Enqueue(analytics.Track{
						UserId:    "some-user-uuid",
						Event:     "app created",
						Timestamp: time.Date(2018, 8, 9, 8, 8, 8, 0, time.UTC),
						Properties: map[string]interface{}{
							"buildpack":      "ruby",
							"os":             runtime.GOOS,
							"plugin_version": "some-version",
							"os_version":     "some-os-version",
							"proxy":          "false",
						},
					})

					mockAnalytics.EXPECT().Enqueue(analytics.Track{
						UserId:    "some-user-uuid",
						Event:     "app created",
						Timestamp: time.Date(2018, 8, 8, 9, 7, 8, 0, time.UTC),
						Properties: map[string]interface{}{
							"buildpack":      "go",
							"os":             runtime.GOOS,
							"plugin_version": "some-version",
							"os_version":     "some-os-version",
							"proxy":          "false",
						},
					})

					startDaemon()
					<-time.After(200 * time.Millisecond)
					aDaemon.Stop()
				})

				It("sends the events and continues polling", func() {
					mockAnalytics.EXPECT().Enqueue(analytics.Track{
						UserId:    "some-user-uuid",
						Event:     "app created",
						Timestamp: time.Date(2018, 8, 9, 8, 8, 8, 0, time.UTC),
						Properties: map[string]interface{}{
							"buildpack":      "ruby",
							"os":             runtime.GOOS,
							"plugin_version": "some-version",
							"os_version":     "some-os-version",
							"proxy":          "false",
						},
					})

					mockAnalytics.EXPECT().Enqueue(analytics.Track{
						UserId:    "some-user-uuid",
						Event:     "app created",
						Timestamp: time.Date(2018, 8, 8, 9, 7, 8, 0, time.UTC),
						Properties: map[string]interface{}{
							"buildpack":      "go",
							"os":             runtime.GOOS,
							"plugin_version": "some-version",
							"os_version":     "some-os-version",
							"proxy":          "false",
						},
					})

					mockAnalytics.EXPECT().Enqueue(analytics.Track{
						UserId:    "some-user-uuid",
						Event:     "app created",
						Timestamp: time.Date(2018, 8, 10, 8, 8, 8, 0, time.UTC),
						Properties: map[string]interface{}{
							"buildpack":      "java",
							"os":             runtime.GOOS,
							"plugin_version": "some-version",
							"os_version":     "some-os-version",
							"proxy":          "false",
						},
					})

					mockAnalytics.EXPECT().Enqueue(analytics.Track{
						UserId:    "some-user-uuid",
						Event:     "app created",
						Timestamp: time.Date(2018, 8, 11, 8, 8, 8, 0, time.UTC),
						Properties: map[string]interface{}{
							"buildpack":      "nodejs",
							"os":             runtime.GOOS,
							"plugin_version": "some-version",
							"os_version":     "some-os-version",
							"proxy":          "false",
						},
					})
					startDaemon()
					<-time.After(2030 * time.Millisecond)
					aDaemon.Stop()
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
			Context("when there are subsequent app push events", func() {
				BeforeEach(func() {
					ccServer.AppendHandlers(ghttp.CombineHandlers(
						ghttp.VerifyRequest(http.MethodGet, "/v2/events"),
						ghttp.RespondWith(http.StatusOK, fakeResponse([]string{
							fakePushEvent("2018-08-09T08:08:08Z", "ruby_buildpack"),
							fakePushEvent("2018-08-08T09:07:08Z", "go_buildpack"),
						})),
					),
						ghttp.CombineHandlers(
							ghttp.VerifyRequest(http.MethodGet, "/v2/events"),
							ghttp.RespondWith(http.StatusOK, fakeResponse([]string{})),
						))
				})
				It("sends the subsequent app push events", func() {
					mockAnalytics.EXPECT().Enqueue(analytics.Track{
						UserId:    "some-user-uuid",
						Event:     "app created",
						Timestamp: time.Date(2018, 8, 9, 8, 8, 8, 0, time.UTC),
						Properties: map[string]interface{}{
							"buildpack":      "ruby",
							"os":             runtime.GOOS,
							"plugin_version": "some-version",
							"os_version":     "some-os-version",
							"proxy":          "false",
						},
					})

					mockAnalytics.EXPECT().Enqueue(analytics.Track{
						UserId:    "some-user-uuid",
						Event:     "app created",
						Timestamp: time.Date(2018, 8, 8, 9, 7, 8, 0, time.UTC),
						Properties: map[string]interface{}{
							"buildpack":      "go",
							"os":             runtime.GOOS,
							"plugin_version": "some-version",
							"os_version":     "some-os-version",
							"proxy":          "false",
						},
					})

					startDaemon()
					<-time.After(1030 * time.Millisecond)
					aDaemon.Stop()
				})
			})

			Context("when there is a subsequent app crash event", func() {
				BeforeEach(func() {
					ccServer.AppendHandlers(ghttp.CombineHandlers(
						ghttp.VerifyRequest(http.MethodGet, "/v2/events"),
						ghttp.RespondWith(http.StatusOK, fakeResponse([]string{
							fakeCrashEvent("2018-08-09T08:08:08Z"),
						})),
					),
						ghttp.CombineHandlers(
							ghttp.VerifyRequest(http.MethodGet, "/v2/events"),
							ghttp.RespondWith(http.StatusOK, fakeResponse([]string{})),
						))
				})
				It("sends the crash event", func() {
					mockAnalytics.EXPECT().Enqueue(analytics.Track{
						UserId:    "some-user-uuid",
						Event:     "app push failed",
						Timestamp: time.Date(2018, 8, 9, 8, 8, 8, 0, time.UTC),
						Properties: map[string]interface{}{
							"os":             runtime.GOOS,
							"plugin_version": "some-version",
							"os_version":     "some-os-version",
							"proxy":          "false",
						},
					})

					startDaemon()
					<-time.After(1030 * time.Millisecond)
					aDaemon.Stop()
				})
			})

			Context("when there is a subsequent service create event", func() {
				BeforeEach(func() {
					ccServer.AppendHandlers(ghttp.CombineHandlers(
						ghttp.VerifyRequest(http.MethodGet, "/v2/events"),
						ghttp.RespondWith(http.StatusOK, fakeResponse([]string{
							fakeServiceCreateEvent("2018-08-09T08:08:08Z", "some-service-plan-guid"),
						})),
					))

					ccServer.AppendHandlers(ghttp.CombineHandlers(
						ghttp.VerifyRequest(http.MethodGet, "/v2/service_plans/some-service-plan-guid"),
						ghttp.RespondWith(http.StatusOK, fakeUrlResponse("/some-service-url")),
					))

					ccServer.AppendHandlers(ghttp.CombineHandlers(
						ghttp.VerifyRequest(http.MethodGet, "/some-service-url"),
						ghttp.RespondWith(http.StatusOK, fakeLabelResponse("p-circuit-breaker-dashboard")),
					))

					ccServer.AppendHandlers(ghttp.CombineHandlers(
						ghttp.VerifyRequest(http.MethodGet, "/v2/events"),
						ghttp.RespondWith(http.StatusOK, fakeResponse([]string{})),
					))
				})
				It("sends the service create event", func() {
					mockAnalytics.EXPECT().Enqueue(analytics.Track{
						UserId:    "some-user-uuid",
						Event:     "created service",
						Timestamp: time.Date(2018, 8, 9, 8, 8, 8, 0, time.UTC),
						Properties: map[string]interface{}{
							"service":        "p-circuit-breaker-dashboard",
							"os":             runtime.GOOS,
							"plugin_version": "some-version",
							"os_version":     "some-os-version",
							"proxy":          "false",
						},
					})

					startDaemon()
					<-time.After(1030 * time.Millisecond)
					aDaemon.Stop()
				})
			})

			Context("when there is a subsequent service bind event", func() {
				BeforeEach(func() {
					ccServer.AppendHandlers(ghttp.CombineHandlers(
						ghttp.VerifyRequest(http.MethodGet, "/v2/events"),
						ghttp.RespondWith(http.StatusOK, fakeResponse([]string{
							fakeServiceBindEvent("2018-08-09T08:08:08Z", "some-guid"),
						})),
					))

					ccServer.AppendHandlers(ghttp.CombineHandlers(
						ghttp.VerifyRequest(http.MethodGet, "/v2/service_instances/some-guid"),
						ghttp.RespondWith(http.StatusOK, fakeUrlResponse("/some-service-url")),
					))

					ccServer.AppendHandlers(ghttp.CombineHandlers(
						ghttp.VerifyRequest(http.MethodGet, "/some-service-url"),
						ghttp.RespondWith(http.StatusOK, fakeLabelResponse("p-circuit-breaker-dashboard")),
					))

					ccServer.AppendHandlers(ghttp.CombineHandlers(
						ghttp.VerifyRequest(http.MethodGet, "/v2/events"),
						ghttp.RespondWith(http.StatusOK, fakeResponse([]string{})),
					))
				})
				It("sends the service bind event", func() {
					mockAnalytics.EXPECT().Enqueue(analytics.Track{
						UserId:    "some-user-uuid",
						Event:     "app bound to service",
						Timestamp: time.Date(2018, 8, 9, 8, 8, 8, 0, time.UTC),
						Properties: map[string]interface{}{
							"service":        "p-circuit-breaker-dashboard",
							"os":             runtime.GOOS,
							"plugin_version": "some-version",
							"os_version":     "some-os-version",
							"proxy":          "false",
						},
					})

					startDaemon()
					<-time.After(1030 * time.Millisecond)
					aDaemon.Stop()
				})
			})

			Context("when there is a subsequent restage event", func() {
				BeforeEach(func() {
					ccServer.AppendHandlers(ghttp.CombineHandlers(
						ghttp.VerifyRequest(http.MethodGet, "/v2/events"),
						ghttp.RespondWith(http.StatusOK, fakeResponse([]string{fakeRestageEvent("2018-08-08T08:08:08Z")})),
					))

					ccServer.AppendHandlers(ghttp.CombineHandlers(
						ghttp.VerifyRequest(http.MethodGet, "/v2/events"),
						ghttp.RespondWith(http.StatusOK, fakeResponse([]string{})),
					))
				})
				It("sends the restage event", func() {
					mockAnalytics.EXPECT().Enqueue(analytics.Track{
						UserId:    "some-user-uuid",
						Event:     "app restage",
						Timestamp: time.Date(2018, 8, 8, 8, 8, 8, 0, time.UTC),
						Properties: map[string]interface{}{
							"os":             runtime.GOOS,
							"plugin_version": "some-version",
							"os_version":     "some-os-version",
							"proxy":          "false",
						},
					})

					startDaemon()
					<-time.After(1030 * time.Millisecond)
					aDaemon.Stop()
				})
			})

			Context("when there is a subsequent user-provided-service event", func() {
				BeforeEach(func() {
					ccServer.AppendHandlers(ghttp.CombineHandlers(
						ghttp.VerifyRequest(http.MethodGet, "/v2/events"),
						ghttp.RespondWith(http.StatusOK, fakeResponse([]string{
							fakeUserProvidedServiceEvent("2018-08-08T08:08:08Z"),
						})),
					))

					ccServer.AppendHandlers(ghttp.CombineHandlers(
						ghttp.VerifyRequest(http.MethodGet, "/v2/events"),
						ghttp.RespondWith(http.StatusOK, fakeResponse([]string{})),
					))
				})
				It("sends the user-provided-service event", func() {
					mockAnalytics.EXPECT().Enqueue(analytics.Track{
						UserId:    "some-user-uuid",
						Event:     "created user provided service",
						Timestamp: time.Date(2018, 8, 8, 8, 8, 8, 0, time.UTC),
						Properties: map[string]interface{}{
							"os":             runtime.GOOS,
							"plugin_version": "some-version",
							"os_version":     "some-os-version",
							"proxy":          "false",
						},
					})

					startDaemon()
					<-time.After(1030 * time.Millisecond)
					aDaemon.Stop()
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
					aDaemon.Stop()
				})
			})
		})
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

var restageAppEventTemplate = `
{
	"entity": {
		"type": "audit.app.restage",
		"timestamp": "%s",
		"metadata": {}
	}
}
`

var serviceBindEventTemplate = `
{
	"entity": {
		"type": "audit.service_binding.create",
		"timestamp": "%s",
		"metadata": {
			"request": {
				"relationships": {
					"service_instance": {
						"data": {
							"guid": "%s"
						}	
					}
				}
			}
		}
	}
}
`

var serviceCreateEventTemplate = `
{
	"entity": {
		"type": "audit.service_instance.create",
		"timestamp": "%s",
		"metadata": {
			"request": {
				"service_plan_guid": "%s"
			}
		}
	}
}
`
var userProvidedServiceEventTemplate = `
{
	"entity": {
		"type": "audit.user_provided_service_instance.create",
		"timestamp": "%s",
		"metadata": {
		}
	}
}
`

var crashAppEventTemplate = `
{
	"entity": {
		"type": "app.crash",
		"timestamp": "%s",
		"metadata": { 
		}
	}
}
`

var urlResponseTemplate = `
{
	"entity": {
		"service_url": "%s"
	}
}
`

var labelResponseTemplate = `
{
	"entity": {
		"label": "%s"
	}
}
`

func fakeEvent(eventType, timestamp, buildpack string) string {
	return fmt.Sprintf(pushAppEventTemplate, eventType, timestamp, buildpack)
}

func fakePushEvent(timestamp, buildpack string) string {
	return fakeEvent("audit.app.create", timestamp, buildpack)
}

func fakeServiceCreateEvent(timestamp, servicePlanGUID string) string {
	return fmt.Sprintf(serviceCreateEventTemplate, timestamp, servicePlanGUID)
}

func fakeServiceBindEvent(timestamp, guid string) string {
	return fmt.Sprintf(serviceBindEventTemplate, timestamp, guid)
}

func fakeRestageEvent(timestamp string) string {
	return fmt.Sprintf(restageAppEventTemplate, timestamp)
}

func fakeCrashEvent(timestamp string) string {
	return fmt.Sprintf(crashAppEventTemplate, timestamp)
}

func fakeUrlResponse(serviceURL string) string {
	return fmt.Sprintf(urlResponseTemplate, serviceURL)
}

func fakeLabelResponse(label string) string {
	return fmt.Sprintf(labelResponseTemplate, label)
}

func fakeUserProvidedServiceEvent(timestamp string) string {
	return fmt.Sprintf(userProvidedServiceEventTemplate, timestamp)
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
