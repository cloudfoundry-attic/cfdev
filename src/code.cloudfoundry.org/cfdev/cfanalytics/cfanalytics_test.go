package cfanalytics_test

import (
	"io/ioutil"
	"os"

	"path"

	"bytes"
	"net"

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
		cfanalytics.PromptOptIn(conf, &mockUI)

		Expect(analyticsFilePath).To(BeAnExistingFile())
		Expect(mockUI.WasCalled).To(Equal(true))
	})

	It("prompts user when analytics file is present but empty", func() {
		ioutil.WriteFile(analyticsFilePath, []byte(""), 0777)

		cfanalytics.PromptOptIn(conf, &mockUI)

		Expect(analyticsFilePath).To(BeAnExistingFile())
		Expect(mockUI.WasCalled).To(Equal(true))
	})

	It("does not prompt user when analytics file is present & contains opted in", func() {
		os.MkdirAll(conf.AnalyticsDir, 0755)
		ioutil.WriteFile(analyticsFilePath, []byte("optedin"), 0755)

		cfanalytics.PromptOptIn(conf, &mockUI)

		Expect(analyticsFilePath).To(BeAnExistingFile())
		Expect(mockUI.WasCalled).To(Equal(false))
	})

	It("does not prompt user when analytics file is present & contains opted out", func() {
		os.MkdirAll(conf.AnalyticsDir, 0755)
		ioutil.WriteFile(analyticsFilePath, []byte("optedout"), 0755)

		Expect(analyticsFilePath).To(BeAnExistingFile())
		Expect(mockUI.WasCalled).To(Equal(false))
	})

	Context("User responds to prompt with", func() {
		BeforeEach(func() {
			os.MkdirAll(conf.AnalyticsDir, 0755)
			ioutil.WriteFile(analyticsFilePath, []byte(""), 0755)
		})

		It("anything & analytics file does not exist", func() {
			os.Remove(analyticsFilePath)
			err := cfanalytics.SetTelemetryState("y", conf)
			Expect(err).To(HaveOccurred())
		})

		It("yes & analytics file will contain optin", func() {
			err := cfanalytics.SetTelemetryState("yes", conf)
			Expect(err).ToNot(HaveOccurred())

			content, _ := ioutil.ReadFile(analyticsFilePath)
			Expect(string(content)).To(Equal("optin"))
		})

		It("Y & analytics file will contain optin", func() {
			err := cfanalytics.SetTelemetryState("Y", conf)
			Expect(err).ToNot(HaveOccurred())

			content, _ := ioutil.ReadFile(analyticsFilePath)
			Expect(string(content)).To(Equal("optin"))
		})

		It("no & analytics file will contain optout", func() {
			err := cfanalytics.SetTelemetryState("no", conf)
			Expect(err).ToNot(HaveOccurred())

			content, _ := ioutil.ReadFile(analyticsFilePath)
			Expect(string(content)).To(Equal("optout"))
		})

		It("N & analytics file will contain optout", func() {
			err := cfanalytics.SetTelemetryState("N", conf)
			Expect(err).ToNot(HaveOccurred())

			content, _ := ioutil.ReadFile(analyticsFilePath)
			Expect(string(content)).To(Equal("optout"))
		})

		It("MumboJumbo & analytics file will contain optout", func() {
			err := cfanalytics.SetTelemetryState("MumboJumbo", conf)
			Expect(err).ToNot(HaveOccurred())

			content, _ := ioutil.ReadFile(analyticsFilePath)
			Expect(string(content)).To(Equal("optout"))
		})

		It("yes & analytics file is overwritten", func() {
			ioutil.WriteFile(analyticsFilePath, []byte("JUNK"), 0755)
			content, _ := ioutil.ReadFile(analyticsFilePath)
			Expect(string(content[:])).To(Equal("JUNK"))

			err := cfanalytics.SetTelemetryState("yes", conf)
			Expect(err).ToNot(HaveOccurred())

			content, _ = ioutil.ReadFile(analyticsFilePath)
			Expect(string(content)).To(Equal("optin"))
		})
	})

	Context("GetUUID", func() {
		It("should return non-empty string", func() {
			Expect(cfanalytics.GetUUID()).ToNot(BeEmpty())
		})

		It("should not return unhashed mac address", func() {
			var addr string
			interfaces, err := net.Interfaces()
			if err == nil {
				for _, i := range interfaces {
					if i.Flags&net.FlagUp != 0 && bytes.Compare(i.HardwareAddr, nil) != 0 {
						addr = i.HardwareAddr.String()
						break
					}
				}
			}

			Expect(cfanalytics.GetUUID()).ToNot(Equal(addr))
		})
	})

	Context("TrackEvent", func() {
		It("should track event", func() {
			mockClient := MockClient{
				WasCalledWith: analytics.Track{},
			}

			err := cfanalytics.TrackEvent("TEST EVENT", "cf", &mockClient)
			Expect(err).ToNot(HaveOccurred())

			out, err := json.Marshal(mockClient.WasCalledWith)
			Expect(err).ToNot(HaveOccurred())

			Expect(out).Should(ContainSubstring("TEST EVENT"))
		})
	})
})
