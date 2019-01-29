package bosh_test

import (
	"code.cloudfoundry.org/cfdev/cfanalytics"
	"code.cloudfoundry.org/cfdev/config"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	cmd "code.cloudfoundry.org/cfdev/cmd/bosh"
	"code.cloudfoundry.org/cfdev/cmd/bosh/mocks"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Bosh", func() {
	var (
		mockController      *gomock.Controller
		mockUI              *mocks.MockUI
		tmpDir              string
		boshCmd             *cmd.Bosh
		mockAnalyticsClient *mocks.MockAnalyticsClient
	)

	BeforeEach(func() {
		mockController = gomock.NewController(GinkgoT())
		mockAnalyticsClient = mocks.NewMockAnalyticsClient(mockController)
		mockUI = mocks.NewMockUI(mockController)

		cfg := config.Config{
			StateBosh: tmpDir,
		}

		content := `---
BOSH_ENVIRONMENT: 10.0.0.1
BOSH_CLIENT: admin`

		ioutil.WriteFile(filepath.Join(tmpDir, "env.yml"), []byte(content), 0600)

		boshCmd = &cmd.Bosh{
			UI:          mockUI,
			Analytics:   mockAnalyticsClient,
			Config: cfg,
		}
	})

	AfterEach(func() {
		mockController.Finish()
		os.RemoveAll(tmpDir)
	})

	Describe("Env", func() {
		Context("when the environment has BOSH_* env vars set", func() {
			BeforeEach(func() {
				for _, envvar := range os.Environ() {
					if strings.HasPrefix(envvar, "BOSH_") {
						key := strings.Split(envvar, "=")[0]
						os.Unsetenv(key)
					}
				}
				os.Setenv("BOSH_SOME_VAR", "some-val")
				os.Setenv("BOSH_SOME_OTHER_VAR", "some-other-val")
			})

			It("prints unset and export statements", func() {
				mockAnalyticsClient.EXPECT().Event(cfanalytics.BOSH_ENV)
				mockUI.EXPECT().Say(gomock.Any()).Do(func(arg string) {
					Expect(arg).To(ContainSubstring(`Remove-Item Env:BOSH_SOME_OTHER_VAR;`))
					Expect(arg).To(ContainSubstring(`Remove-Item Env:BOSH_SOME_VAR;`))
					Expect(arg).To(ContainSubstring(`$env:BOSH_ENVIRONMENT="10.0.0.1";`))
					Expect(arg).To(ContainSubstring(`$env:BOSH_CLIENT="admin";`))
				})

				Expect(boshCmd.Env()).To(Succeed())
			})
		})
	})
})
