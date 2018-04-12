package cmd_test

import (
	"io/ioutil"
	"os"
	"path"

	"code.cloudfoundry.org/cfdev/cmd"
	"code.cloudfoundry.org/cfdev/config"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

type MockUI struct {
	WasCalledWith string
}

func (m *MockUI) Say(message string, args ...interface{}) {
	m.WasCalledWith = message
}

var _ = Describe("Telemetry", func() {
	var (
		tmpDir            string
		mockUI            MockUI
		conf              config.Config
		analyticsFilePath string
		telCmd            cmd.Telemetry
	)

	BeforeEach(func() {
		mockUI = MockUI{
			WasCalledWith: "",
		}

		tmpDir, _ = ioutil.TempDir(os.TempDir(), "testdir")
		conf = config.Config{
			AnalyticsDir:  path.Join(tmpDir, "analytics"),
			AnalyticsFile: "analytics-text.txt",
		}

		analyticsFilePath = path.Join(conf.AnalyticsDir, conf.AnalyticsFile)

		telCmd = cmd.Telemetry{
			Config: conf,
			UI:     &mockUI,
		}

		os.MkdirAll(conf.AnalyticsDir, 0755)
	})

	AfterEach(func() {
		os.RemoveAll(tmpDir)
	})

	Context("first arg", func() {
		It("ON", func() {
			ioutil.WriteFile(analyticsFilePath, []byte(""), 0755)

			telCmd.Run([]string{"oN"})

			contents, err := ioutil.ReadFile(analyticsFilePath)
			Expect(err).NotTo(HaveOccurred())
			Expect(string(contents[:])).To(Equal("optin"))
			Expect(mockUI.WasCalledWith).To(Equal("Telemetry is turned ON"))
		})

		It("OFF", func() {
			ioutil.WriteFile(analyticsFilePath, []byte(""), 0755)

			telCmd.Run([]string{"oFf"})

			contents, err := ioutil.ReadFile(analyticsFilePath)
			Expect(err).NotTo(HaveOccurred())
			Expect(string(contents[:])).To(Equal("optout"))
			Expect(mockUI.WasCalledWith).To(Equal("Telemetry is turned OFF"))
		})
	})

	Context("No args displays status", func() {

		It("ON", func() {
			ioutil.WriteFile(analyticsFilePath, []byte("optin"), 0755)

			telCmd.Run([]string{})

			Expect(mockUI.WasCalledWith).To(Equal("Telemetry is turned ON"))
		})

		It("OFF", func() {
			ioutil.WriteFile(analyticsFilePath, []byte("optout"), 0755)

			telCmd.Run([]string{})

			Expect(mockUI.WasCalledWith).To(Equal("Telemetry is turned OFF"))
		})
	})
})
