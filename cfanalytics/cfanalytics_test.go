package cfanalytics_test

import (
	"code.cloudfoundry.org/cfdev/cfanalytics"
	"code.cloudfoundry.org/cfdev/cfanalytics/mocks"
	"github.com/denisbrodbeck/machineid"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gstruct"
	"gopkg.in/segmentio/analytics-go.v3"
	"runtime"
	"time"
)

var _ = Describe("Analytics", func() {
	var (
		mockController *gomock.Controller
		mockClient     *mocks.MockClient
		mockToggle     *mocks.MockToggle
		mockUI         *mocks.MockUI
		exitChan       chan struct{}
		subject        *cfanalytics.Analytics
	)
	BeforeEach(func() {
		mockController = gomock.NewController(GinkgoT())
		mockClient = mocks.NewMockClient(mockController)
		mockToggle = mocks.NewMockToggle(mockController)
		mockUI = mocks.NewMockUI(mockController)
		exitChan = make(chan struct{}, 1)
		subject = cfanalytics.New(mockToggle, mockClient, "4.5.6-unit-test", "some-os-version", "false", exitChan, mockUI)
	})
	AfterEach(func() {
		mockController.Finish()
	})

	Describe("PromptOptInIfNeeded with empty message", func() {
		Context("When user has NOT yet answered optin prompt", func() {
			BeforeEach(func() {
				mockToggle.EXPECT().Defined().Return(false).AnyTimes()
			})
			It("prompts user", func() {
				mockToggle.EXPECT().SetCFAnalyticsEnabled(gomock.Any()).AnyTimes()
				mockUI.EXPECT().Ask(gomock.Any()).Do(func(prompt string) {
					Expect(prompt).To(ContainSubstring("Are you ok with CF Dev periodically capturing anonymized telemetry [y/N]?"))
				})
				Expect(subject.PromptOptInIfNeeded("")).To(Succeed())
			})
			for _, answer := range []string{"yes", "y", "yEs"} {
				Context("user answers "+answer, func() {
					BeforeEach(func() { mockUI.EXPECT().Ask(gomock.Any()).Return(answer) })
					It("saves optin", func() {
						mockToggle.EXPECT().SetCFAnalyticsEnabled(true)

						Expect(subject.PromptOptInIfNeeded("")).To(Succeed())
					})
				})
			}
			for _, answer := range []string{"no", "N", "anything", ""} {
				Context("user answers "+answer, func() {
					BeforeEach(func() { mockUI.EXPECT().Ask(gomock.Any()).Return(answer) })
					It("saves optout", func() {
						mockToggle.EXPECT().SetCFAnalyticsEnabled(false)

						Expect(subject.PromptOptInIfNeeded("")).To(Succeed())
					})
				})
			}
			Context("user hits ctrl-c", func() {
				BeforeEach(func() {
					mockUI.EXPECT().Ask(gomock.Any()).Return("")
					exitChan <- struct{}{}
				})
				It("does not write set a value on toggle", func() {
					Expect(subject.PromptOptInIfNeeded("")).To(MatchError("Exit while waiting for telemetry prompt"))
				})
			})
		})
		Context("When user has answered optin prompt", func() {
			BeforeEach(func() {
				mockToggle.EXPECT().Defined().AnyTimes().Return(true)
			})
			It("does not ask again", func() {
				Expect(subject.PromptOptInIfNeeded("")).To(Succeed())
			})
		})
	})
	Describe("PromptOptInIfNeeded with custom message", func() {
		Context("When user has NOT yet answered any optin prompt at all", func() {
			BeforeEach(func() {
				mockToggle.EXPECT().Defined().Return(false).AnyTimes()
				mockToggle.EXPECT().CustomAnalyticsDefined().Return(false).AnyTimes()
			})
			It("prompts user", func() {
				mockToggle.EXPECT().SetCustomAnalyticsEnabled(gomock.Any()).AnyTimes()
				mockUI.EXPECT().Ask(gomock.Any()).Do(func(prompt string) {
					Expect(prompt).To(ContainSubstring("some-custom-message"))
				})
				Expect(subject.PromptOptInIfNeeded("some-custom-message")).To(Succeed())
			})
			for _, answer := range []string{"yes", "y", "yEs"} {
				Context("user answers "+answer, func() {
					BeforeEach(func() { mockUI.EXPECT().Ask(gomock.Any()).Return(answer) })
					It("saves optin", func() {
						mockToggle.EXPECT().SetCustomAnalyticsEnabled(true)

						Expect(subject.PromptOptInIfNeeded("some-custom-message")).To(Succeed())
					})
				})
			}
			for _, answer := range []string{"no", "N", "anything", ""} {
				Context("user answers "+answer, func() {
					BeforeEach(func() { mockUI.EXPECT().Ask(gomock.Any()).Return(answer) })
					It("saves optout", func() {
						mockToggle.EXPECT().SetCustomAnalyticsEnabled(false)

						Expect(subject.PromptOptInIfNeeded("some-custom-message")).To(Succeed())
					})
				})
			}
			Context("user hits ctrl-c", func() {
				BeforeEach(func() {
					mockUI.EXPECT().Ask(gomock.Any()).Return("")
					exitChan <- struct{}{}
				})
				It("does not write set a value on toggle", func() {
					Expect(subject.PromptOptInIfNeeded("some-custom-message")).To(MatchError("Exit while waiting for telemetry prompt"))
				})
			})
		})
		Context("When user has answered custom optin prompt already", func() {
			BeforeEach(func() {
				mockToggle.EXPECT().Defined().AnyTimes().Return(true)
				mockToggle.EXPECT().CustomAnalyticsDefined().Return(true).AnyTimes()
			})
			It("does not ask again", func() {
				Expect(subject.PromptOptInIfNeeded("some-custom-message")).To(Succeed())
			})
		})
		Context("When user has answered standard optin prompt but not custom prompt", func() {
			BeforeEach(func() {
				mockToggle.EXPECT().Defined().AnyTimes().Return(true)
				mockToggle.EXPECT().CustomAnalyticsDefined().Return(false).AnyTimes()
			})
			It("prompts user", func() {
				mockToggle.EXPECT().SetCustomAnalyticsEnabled(gomock.Any()).AnyTimes()
				mockUI.EXPECT().Ask(gomock.Any()).Do(func(prompt string) {
					Expect(prompt).To(ContainSubstring("some-custom-message"))
				})
				Expect(subject.PromptOptInIfNeeded("some-custom-message")).To(Succeed())
			})
			for _, answer := range []string{"yes", "y", "yEs"} {
				Context("user answers "+answer, func() {
					BeforeEach(func() { mockUI.EXPECT().Ask(gomock.Any()).Return(answer) })
					It("saves optin", func() {
						mockToggle.EXPECT().SetCustomAnalyticsEnabled(true)

						Expect(subject.PromptOptInIfNeeded("some-custom-message")).To(Succeed())
					})
				})
			}
			for _, answer := range []string{"no", "N", "anything", ""} {
				Context("user answers "+answer, func() {
					BeforeEach(func() { mockUI.EXPECT().Ask(gomock.Any()).Return(answer) })
					It("saves optout", func() {
						mockToggle.EXPECT().SetCustomAnalyticsEnabled(false)

						Expect(subject.PromptOptInIfNeeded("some-custom-message")).To(Succeed())
					})
				})
			}
			Context("user hits ctrl-c", func() {
				BeforeEach(func() {
					mockUI.EXPECT().Ask(gomock.Any()).Return("")
					exitChan <- struct{}{}
				})
				It("does not write set a value on toggle", func() {
					Expect(subject.PromptOptInIfNeeded("some-custom-message")).To(MatchError("Exit while waiting for telemetry prompt"))
				})
			})
		})
	})
	Describe("Event", func() {
		Context("opt out", func() {
			BeforeEach(func() {
				mockToggle.EXPECT().Enabled().AnyTimes().Return(false)
			})
			It("does nothing and succeeds", func() {
				Expect(subject.Event("anevent", map[string]interface{}{"mykey": "myval"})).To(Succeed())
			})
		})
		Context("opt in", func() {
			BeforeEach(func() {
				mockToggle.EXPECT().Enabled().AnyTimes().Return(true)
				mockToggle.EXPECT().GetProps().AnyTimes().Return(map[string]interface{}{
					"type": "cf.1.2.3.iso",
				})
			})
			It("sends identity and event to segmentio", func() {
				uuid, _ := machineid.ProtectedID("cfdev")

				mockClient.EXPECT().Enqueue(gomock.Any()).Do(func(msg analytics.Message) {
					Expect(msg).To(Equal(analytics.Identify{
						UserId: uuid,
					}))
				})
				mockClient.EXPECT().Enqueue(gomock.Any()).Do(func(msg analytics.Message) {
					Expect(msg).To(gstruct.MatchFields(gstruct.IgnoreExtras, gstruct.Fields{
						"UserId":    Equal(uuid),
						"Event":     Equal("anevent"),
						"Timestamp": BeTemporally(">=", time.Now().Add(-1*time.Minute)),
						"Properties": BeEquivalentTo(map[string]interface{}{
							"os":             runtime.GOOS,
							"plugin_version": "4.5.6-unit-test",
							"os_version":     "some-os-version",
							"proxy":          "false",
							"type":           "cf.1.2.3.iso",
							"mykey":          "myval",
						}),
					}))
				})

				Expect(subject.Event("anevent", map[string]interface{}{"mykey": "myval"})).To(Succeed())
			})
		})
	})
})
