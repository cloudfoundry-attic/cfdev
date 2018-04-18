package cfanalytics_test

import (
	"io/ioutil"
	"os"

	"path"

	"encoding/json"

	"code.cloudfoundry.org/cfdev/cfanalytics"
	"code.cloudfoundry.org/cfdev/config"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"gopkg.in/segmentio/analytics-go.v3"
)

type MockUI struct {
	WasCalled bool
}

type UI interface {
	Say(message string, args ...interface{})
	Ask(prompt string) (answer string)
}

func (m *MockUI) Ask(prompt string) (answer string) {
	if prompt == `
CF Dev collects anonymous usage data to help us improve your user experience. We intend to share these anonymous usage analytics with user community by publishing quarterly reports at :

https://github.com/pivotal-cf/cfdev/wiki/Telemetry

Are you ok with CF Dev periodically capturing anonymized telemetry [y/N]?` {
		m.WasCalled = true
	}

	return "yes"
}

func (m *MockUI) Say(message string, args ...interface{}) {
}

var _ = Describe("Optin", func() {
	var (
		tmpDir            string
		mockUI            MockUI
		conf              config.Config
		analyticsFilePath string
	)

	BeforeEach(func() {
		mockUI = MockUI{
			WasCalled: false,
		}

		tmpDir, _ = ioutil.TempDir(os.TempDir(), "testdir")
		conf = config.Config{
			AnalyticsDir:  path.Join(tmpDir, "analytics"),
			AnalyticsFile: "analytics-text.txt",
		}
		analyticsFilePath = path.Join(conf.AnalyticsDir, conf.AnalyticsFile)
	})

	AfterEach(func() {
		os.RemoveAll(tmpDir)
	})

	It("prompts user & creates file when analytics file is missing", func() {
		Expect(cfanalytics.PromptOptIn(conf, &mockUI)).To(Succeed())

		Expect(analyticsFilePath).To(BeAnExistingFile())
		Expect(mockUI.WasCalled).To(Equal(true))
	})

	It("prompts user when analytics file is present but empty", func() {
		ioutil.WriteFile(analyticsFilePath, []byte(""), 0777)

		Expect(cfanalytics.PromptOptIn(conf, &mockUI)).To(Succeed())

		Expect(analyticsFilePath).To(BeAnExistingFile())
		Expect(mockUI.WasCalled).To(Equal(true))
	})

	It("does not prompt user when analytics file is present & contains opted in", func() {
		os.MkdirAll(conf.AnalyticsDir, 0755)
		ioutil.WriteFile(analyticsFilePath, []byte("optedin"), 0755)

		Expect(cfanalytics.PromptOptIn(conf, &mockUI)).To(Succeed())

		Expect(analyticsFilePath).To(BeAnExistingFile())
		Expect(mockUI.WasCalled).To(Equal(false))
	})

	It("does not prompt user when analytics file is present & contains opted out", func() {
		os.MkdirAll(conf.AnalyticsDir, 0755)
		ioutil.WriteFile(analyticsFilePath, []byte("optedout"), 0755)

		Expect(cfanalytics.PromptOptIn(conf, &mockUI)).To(Succeed())

		Expect(analyticsFilePath).To(BeAnExistingFile())
		Expect(mockUI.WasCalled).To(Equal(false))
	})

	Context("User responds to prompt with", func() {
		It("yes & analytics file will contain optin", func() {
			Expect(cfanalytics.SetTelemetryState("yes", conf)).To(Succeed())
			Expect(ioutil.ReadFile(analyticsFilePath)).To(Equal([]byte("optin")))
		})

		It("Y & analytics file will contain optin", func() {
			Expect(cfanalytics.SetTelemetryState("Y", conf)).To(Succeed())
			Expect(ioutil.ReadFile(analyticsFilePath)).To(Equal([]byte("optin")))
		})

		It("no & analytics file will contain optout", func() {
			Expect(cfanalytics.SetTelemetryState("no", conf)).To(Succeed())
			Expect(ioutil.ReadFile(analyticsFilePath)).To(Equal([]byte("optout")))
		})

		It("N & analytics file will contain optout", func() {
			Expect(cfanalytics.SetTelemetryState("N", conf)).To(Succeed())
			Expect(ioutil.ReadFile(analyticsFilePath)).To(Equal([]byte("optout")))
		})

		It("MumboJumbo & analytics file will contain optout", func() {
			Expect(cfanalytics.SetTelemetryState("MumboJumbo", conf)).To(Succeed())
			Expect(ioutil.ReadFile(analyticsFilePath)).To(Equal([]byte("optout")))
		})

		It("yes & analytics file is overwritten", func() {
			os.MkdirAll(conf.AnalyticsDir, 0755)
			ioutil.WriteFile(analyticsFilePath, []byte("JUNK"), 0755)
			Expect(ioutil.ReadFile(analyticsFilePath)).To(Equal([]byte("JUNK")))

			Expect(cfanalytics.SetTelemetryState("yes", conf)).To(Succeed())

			Expect(ioutil.ReadFile(analyticsFilePath)).To(Equal([]byte("optin")))
		})
	})

	Context("TrackEvent", func() {
		It("should track event", func() {
			mockClient := MockClient{
				WasCalledWith: analytics.Track{},
			}

			err := cfanalytics.TrackEvent("TEST EVENT", map[string]interface{}{"type": "cf"}, &mockClient)
			Expect(err).ToNot(HaveOccurred())

			out, err := json.Marshal(mockClient.WasCalledWith)
			Expect(err).ToNot(HaveOccurred())

			Expect(out).Should(ContainSubstring("TEST EVENT"))
		})
	})
})
