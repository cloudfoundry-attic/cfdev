package cfanalytics_test

import (
	"runtime"
	"time"

	"code.cloudfoundry.org/cfdev/cfanalytics"
	"code.cloudfoundry.org/cfdev/cfanalytics/mocks"
	"github.com/denisbrodbeck/machineid"
	"github.com/golang/mock/gomock"
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
	props     map[string]interface{}
	SetCalled bool
}

func (t *MockToggle) Defined() bool                    { return t.defined }
func (t *MockToggle) Get() bool                        { return t.val }
func (t *MockToggle) Set(v bool) error                 { t.val = v; t.SetCalled = true; return nil }
func (t *MockToggle) GetProps() map[string]interface{} { return t.props }

var _ = Describe("Analytics", func() {
	var (
		mockController *gomock.Controller
		mockClient     *mocks.MockClient

		mockToggle *MockToggle
		exitChan   chan struct{}
		mockUI     MockUI
		subject    *cfanalytics.Analytics
	)
	BeforeEach(func() {
		mockController = gomock.NewController(GinkgoT())
		mockClient = mocks.NewMockClient(mockController)

		mockToggle = &MockToggle{}
		exitChan = make(chan struct{}, 1)
		mockUI = MockUI{WasCalled: false}
		subject = cfanalytics.New(mockToggle, mockClient, "4.5.6-unit-test", exitChan, &mockUI)
	})
	AfterEach(func() {
		mockController.Finish()
	})

	Describe("PromptOptIn", func() {
		Context("When user has NOT yet answered optin prompt", func() {
			BeforeEach(func() { mockToggle.defined = false })
			It("prompts user", func() {
				Expect(subject.PromptOptIn()).To(Succeed())
				Expect(mockUI.WasCalled).To(Equal(true))
			})
			for _, answer := range []string{"yes", "y", "yEs"} {
				Context("user answers "+answer, func() {
					BeforeEach(func() { mockUI.Return = answer })
					It("saves optin", func() {
						Expect(subject.PromptOptIn()).To(Succeed())
						Expect(mockToggle.SetCalled).To(Equal(true))
						Expect(mockToggle.val).To(Equal(true))
					})
				})
			}
			for _, answer := range []string{"no", "N", "anything", ""} {
				Context("user answers "+answer, func() {
					BeforeEach(func() { mockUI.Return = answer })
					It("saves optout", func() {
						Expect(subject.PromptOptIn()).To(Succeed())
						Expect(mockToggle.SetCalled).To(Equal(true))
						Expect(mockToggle.val).To(Equal(false))
					})
				})
			}
			Context("user hits ctrl-c", func() {
				BeforeEach(func() {
					mockUI.Return = ""
					exitChan <- struct{}{}
				})
				It("does not write set a value on toggle", func() {
					Expect(subject.PromptOptIn()).To(MatchError("Exit while waiting for telemetry prompt"))
					Expect(mockToggle.SetCalled).To(Equal(false))
				})
			})
		})
		Context("When user has answered optin prompt", func() {
			BeforeEach(func() { mockToggle.defined = true })
			It("does not ask again", func() {
				Expect(subject.PromptOptIn()).To(Succeed())
				Expect(mockUI.WasCalled).To(Equal(false))
				Expect(mockToggle.SetCalled).To(Equal(false))
			})
		})
	})
	Describe("Event", func() {
		Context("opt out", func() {
			BeforeEach(func() { mockToggle.val = false })
			It("does nothing and succeeds", func() {
				Expect(subject.Event("anevent", map[string]interface{}{"mykey": "myval"})).To(Succeed())
			})
		})
		Context("opt in", func() {
			BeforeEach(func() {
				mockToggle.val = true
				mockToggle.props = map[string]interface{}{
					"type": "cf.1.2.3.iso",
				}
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
							"os":      runtime.GOOS,
							"version": "4.5.6-unit-test",
							"type":    "cf.1.2.3.iso",
							"mykey":   "myval",
						}),
					}))
				})

				Expect(subject.Event("anevent", map[string]interface{}{"mykey": "myval"})).To(Succeed())
			})
		})
	})
})
