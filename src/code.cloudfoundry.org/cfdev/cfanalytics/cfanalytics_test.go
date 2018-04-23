package cfanalytics_test

import (
	"runtime"
	"time"

	"code.cloudfoundry.org/cfdev/cfanalytics"
	"github.com/denisbrodbeck/machineid"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gstruct"
	analytics "gopkg.in/segmentio/analytics-go.v3"
)

type MockUI struct {
	WasCalled bool
	Return    string
}

func (m *MockUI) Ask(prompt string) string {
	if prompt == `
CF Dev collects anonymous usage data to help us improve your user experience. We intend to share these anonymous usage analytics with user community by publishing quarterly reports at :

https://github.com/pivotal-cf/cfdev/wiki/Telemetry

Are you ok with CF Dev periodically capturing anonymized telemetry [y/N]?` {
		m.WasCalled = true
	}

	return m.Return
}

type MockToggle struct {
	defined   bool
	val       bool
	SetCalled bool
}

func (t *MockToggle) Defined() bool    { return t.defined }
func (t *MockToggle) Get() bool        { return t.val }
func (t *MockToggle) Set(v bool) error { t.val = v; t.SetCalled = true; return nil }

type MockClient struct {
	CloseWasCalled    bool
	EnqueueCalledWith analytics.Message
}

func (c *MockClient) Close() error                        { c.CloseWasCalled = true; return nil }
func (c *MockClient) Enqueue(msg analytics.Message) error { c.EnqueueCalledWith = msg; return nil }

var _ = Describe("Analytics", func() {
	var (
		mockToggle *MockToggle
		mockClient *MockClient
		subject    *cfanalytics.Analytics
	)
	BeforeEach(func() {
		mockToggle = &MockToggle{}
		mockClient = &MockClient{}
		subject = cfanalytics.New(mockToggle, mockClient)
	})

	Describe("PromptOptIn", func() {
		var mockUI MockUI
		BeforeEach(func() {
			mockUI = MockUI{
				WasCalled: false,
			}
		})
		Context("When user has not yet answered optin prompt", func() {
			It("prompts user", func() {
				Expect(subject.PromptOptIn(&mockUI)).To(Succeed())
				Expect(mockUI.WasCalled).To(Equal(true))
			})
			for _, answer := range []string{"yes", "y", "yEs"} {
				Context("user answers "+answer, func() {
					BeforeEach(func() { mockUI.Return = answer })
					It("saves optin", func() {
						Expect(subject.PromptOptIn(&mockUI)).To(Succeed())
						Expect(mockToggle.SetCalled).To(Equal(true))
						Expect(mockToggle.val).To(Equal(true))
					})
				})
			}
			for _, answer := range []string{"no", "N", "anything", ""} {
				Context("user answers "+answer, func() {
					BeforeEach(func() { mockUI.Return = answer })
					It("saves optout", func() {
						Expect(subject.PromptOptIn(&mockUI)).To(Succeed())
						Expect(mockToggle.SetCalled).To(Equal(true))
						Expect(mockToggle.val).To(Equal(false))
					})
				})
			}
		})
	})
	Describe("Event", func() {
		Context("opt out", func() {
			BeforeEach(func() { mockToggle.val = false })
			It("does nothing and succeeds", func() {
				Expect(subject.Event("anevent", map[string]interface{}{"mykey": "myval"})).To(Succeed())
				Expect(mockClient.EnqueueCalledWith).To(BeNil())
			})
		})
		Context("opt in", func() {
			BeforeEach(func() { mockToggle.val = true })
			It("sends event to segmentio", func() {
				Expect(subject.Event("anevent", map[string]interface{}{"mykey": "myval"})).To(Succeed())

				uuid, _ := machineid.ProtectedID("cfdev")
				Expect(mockClient.EnqueueCalledWith).To(gstruct.MatchFields(gstruct.IgnoreExtras, gstruct.Fields{
					"UserId":    Equal(uuid),
					"Event":     Equal("anevent"),
					"Timestamp": BeTemporally(">=", time.Now().Add(-1*time.Minute)),
					"Properties": BeEquivalentTo(map[string]interface{}{
						"os":      runtime.GOOS,
						"version": "0.0.2",
						"mykey":   "myval",
					}),
				}))
			})
		})
	})
})
