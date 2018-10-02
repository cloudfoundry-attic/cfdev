package telemetry_test

import (
	"fmt"

	"code.cloudfoundry.org/cfdev/cmd/telemetry"
	"code.cloudfoundry.org/cfdev/cmd/telemetry/mocks"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/spf13/cobra"
)

type MockUI struct {
	WasCalledWith string
}

func (m *MockUI) Say(message string, args ...interface{}) {
	m.WasCalledWith = fmt.Sprintf(message, args...)
}

type MockToggle struct {
	val bool
}

func (t *MockToggle) Get() bool        { return t.val }
func (t *MockToggle) Set(v bool) error { t.val = v; return nil }

var _ = Describe("Telemetry", func() {
	var (
		mockController *gomock.Controller
		mockAnalyticsD *mocks.MockAnalyticsD
		mockUI         MockUI
		mockToggle     *MockToggle
		telCmd         *cobra.Command
	)

	BeforeEach(func() {
		mockController = gomock.NewController(GinkgoT())
		mockAnalyticsD = mocks.NewMockAnalyticsD(mockController)

		mockUI = MockUI{
			WasCalledWith: "",
		}
		mockToggle = &MockToggle{}

		subject := &telemetry.Telemetry{
			UI:              &mockUI,
			AnalyticsToggle: mockToggle,
			AnalyticsD:      mockAnalyticsD,
		}
		telCmd = subject.Cmd()
		telCmd.SetArgs([]string{})
	})

	AfterEach(func() {
		mockController.Finish()
	})

	Context("first arg", func() {
		It("ON", func() {
			mockToggle.val = false

			mockAnalyticsD.EXPECT().IsRunning().Return(false, nil)
			mockAnalyticsD.EXPECT().Start()

			telCmd.SetArgs([]string{"--on"})
			Expect(telCmd.Execute()).To(Succeed())

			Expect(mockToggle.val).To(Equal(true))
			Expect(mockUI.WasCalledWith).To(Equal("Telemetry is turned ON"))
		})

		It("OFF", func() {
			mockToggle.val = true

			mockAnalyticsD.EXPECT().IsRunning().Return(true, nil)
			mockAnalyticsD.EXPECT().Stop()
			mockAnalyticsD.EXPECT().Destroy()

			telCmd.SetArgs([]string{"--off"})
			Expect(telCmd.Execute()).To(Succeed())

			Expect(mockToggle.val).To(Equal(false))
			Expect(mockUI.WasCalledWith).To(Equal("Telemetry is turned OFF"))
		})
	})

	Context("No args displays status", func() {
		It("ON", func() {
			mockToggle.val = true

			Expect(telCmd.Execute()).To(Succeed())

			Expect(mockUI.WasCalledWith).To(Equal("Telemetry is turned ON"))
		})

		It("OFF", func() {
			mockToggle.val = false

			Expect(telCmd.Execute()).To(Succeed())

			Expect(mockUI.WasCalledWith).To(Equal("Telemetry is turned OFF"))
		})
	})
})
