package telemetry_test

import (
	"code.cloudfoundry.org/cfdev/cfanalytics/toggle"
	"code.cloudfoundry.org/cfdev/cmd/telemetry"
	"fmt"
	"io/ioutil"
	"os"

	"code.cloudfoundry.org/cfdev/cmd/telemetry/mocks"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/spf13/cobra"
)

type MockAnalitics struct {
	EventWasCalledWith string
}

func (m *MockAnalitics) Event(event string, data ...map[string]interface{}) error {
	m.EventWasCalledWith = event
	return nil
}

type MockUI struct {
	WasCalledWith string
}

func (m *MockUI) Say(message string, args ...interface{}) {
	m.WasCalledWith = fmt.Sprintf(message, args...)
}

var _ = Describe("Telemetry", func() {
	var (
		mockUI         MockUI
		mockController *gomock.Controller
		mockAnalyticsD *mocks.MockAnalyticsD
		mockAnalytics  MockAnalitics
		t0ggle         *toggle.Toggle
		telCmd         *cobra.Command
		tempFilePath   string
	)

	BeforeEach(func() {
		mockUI = MockUI{}
		mockController = gomock.NewController(GinkgoT())
		mockAnalyticsD = mocks.NewMockAnalyticsD(mockController)
		mockAnalytics = MockAnalitics{}

		tempFile, err := ioutil.TempFile("", "cfdev-telemetry-")
		Expect(err).NotTo(HaveOccurred())
		tempFilePath = tempFile.Name()
	})

	JustBeforeEach(func() {
		t0ggle = toggle.New(tempFilePath)

		subject := &telemetry.Telemetry{
			UI:              &mockUI,
			AnalyticsToggle: t0ggle,
			AnalyticsD:      mockAnalyticsD,
			Analytics:       &mockAnalytics,
		}

		telCmd = subject.Cmd()
		telCmd.SetArgs([]string{})
	})

	AfterEach(func() {
		os.RemoveAll(tempFilePath)
		mockController.Finish()
	})

	Context("when telemetry status is set", func() {
		It("ON", func() {
			mockAnalyticsD.EXPECT().IsRunning().Return(false, nil)
			mockAnalyticsD.EXPECT().Start()

			telCmd.SetArgs([]string{"--on"})
			Expect(telCmd.Execute()).To(Succeed())

			Expect(t0ggle.Enabled()).To(BeTrue())
			Expect(mockUI.WasCalledWith).To(Equal("Telemetry is turned ON"))
			Expect(mockAnalytics.EventWasCalledWith).To(Equal(""))
		})

		It("OFF", func() {
			mockAnalyticsD.EXPECT().IsRunning().Return(true, nil)
			mockAnalyticsD.EXPECT().Stop()
			mockAnalyticsD.EXPECT().Destroy()

			telCmd.SetArgs([]string{"--off"})
			Expect(telCmd.Execute()).To(Succeed())

			Expect(t0ggle.Enabled()).To(BeFalse())
			Expect(mockUI.WasCalledWith).To(Equal("Telemetry is turned OFF"))
			Expect(mockAnalytics.EventWasCalledWith).To(Equal("telemetry off"))
		})
	})

	Describe("telemetry status", func() {
		Context("when cfanalytics is enabled", func() {
			BeforeEach(func() {
				err := ioutil.WriteFile(
					tempFilePath,
					[]byte(`{"cfAnalyticsEnabled": true, "customAnalyticsEnabled": false}`),
					0600)
				Expect(err).NotTo(HaveOccurred())
			})

			It("should display ON", func() {
				Expect(telCmd.Execute()).To(Succeed())

				Expect(mockUI.WasCalledWith).To(Equal("Telemetry is turned ON"))
			})
		})

		Context("when custom analytics is enabled", func() {
			BeforeEach(func() {
				err := ioutil.WriteFile(
					tempFilePath,
					[]byte(`{"cfAnalyticsEnabled": false, "customAnalyticsEnabled": true}`),
					0600)
				Expect(err).NotTo(HaveOccurred())
			})

			It("should display ON", func() {
				Expect(telCmd.Execute()).To(Succeed())

				Expect(mockUI.WasCalledWith).To(Equal("Telemetry is turned ON"))
			})
		})

		Context("when analytics is disabled", func() {
			BeforeEach(func() {
				err := ioutil.WriteFile(
					tempFilePath,
					[]byte(`{"cfAnalyticsEnabled": false, "customAnalyticsEnabled": false}`),
					0600)
				Expect(err).NotTo(HaveOccurred())
			})

			It("should display OFF", func() {
				Expect(telCmd.Execute()).To(Succeed())

				Expect(mockUI.WasCalledWith).To(Equal("Telemetry is turned OFF"))
			})
		})
	})
})
