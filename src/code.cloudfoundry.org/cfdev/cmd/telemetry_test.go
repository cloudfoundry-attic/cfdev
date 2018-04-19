package cmd_test

import (
	"io/ioutil"
	"os"
	"path"

	"code.cloudfoundry.org/cfdev/cmd"
	"code.cloudfoundry.org/cfdev/config"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/spf13/cobra"
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
		telCmd            *cobra.Command
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

		os.MkdirAll(conf.AnalyticsDir, 0755)
		analyticsFilePath = path.Join(conf.AnalyticsDir, conf.AnalyticsFile)

		telCmd = cmd.NewTelemetry(&mockUI, conf)
		telCmd.SetArgs([]string{})
	})

	AfterEach(func() {
		os.RemoveAll(tmpDir)
	})

	Context("first arg", func() {
		It("ON", func() {
			ioutil.WriteFile(analyticsFilePath, []byte(""), 0755)

			telCmd.SetArgs([]string{"--on"})
			Expect(telCmd.Execute()).To(Succeed())

			Expect(ioutil.ReadFile(analyticsFilePath)).To(Equal([]byte("optin")))
			Expect(mockUI.WasCalledWith).To(Equal("Telemetry is turned ON"))
		})

		It("OFF", func() {
			ioutil.WriteFile(analyticsFilePath, []byte(""), 0755)

			telCmd.SetArgs([]string{"--off"})
			Expect(telCmd.Execute()).To(Succeed())

			Expect(ioutil.ReadFile(analyticsFilePath)).To(Equal([]byte("optout")))
			Expect(mockUI.WasCalledWith).To(Equal("Telemetry is turned OFF"))
		})
	})

	Context("No args displays status", func() {
		It("ON", func() {
			ioutil.WriteFile(analyticsFilePath, []byte("optin"), 0755)

			Expect(telCmd.Execute()).To(Succeed())

			Expect(mockUI.WasCalledWith).To(Equal("Telemetry is turned ON"))
		})

		It("OFF", func() {
			ioutil.WriteFile(analyticsFilePath, []byte("optout"), 0755)

			Expect(telCmd.Execute()).To(Succeed())

			Expect(mockUI.WasCalledWith).To(Equal("Telemetry is turned OFF"))
		})
	})
})
