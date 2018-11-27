package bosh_test

import (
	"code.cloudfoundry.org/cfdev/cfanalytics"
	"code.cloudfoundry.org/cfdev/config"
	"fmt"
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
		cfg config.Config
	)

	BeforeEach(func() {
		mockController = gomock.NewController(GinkgoT())
		mockAnalyticsClient = mocks.NewMockAnalyticsClient(mockController)
		mockUI = mocks.NewMockUI(mockController)

		var err error
		tmpDir, err = ioutil.TempDir("", "cmd-bosh-test")
		Expect(err).NotTo(HaveOccurred())

		cfg = config.Config{
			StateBosh: tmpDir,
			BoshDirectorIP: "10.0.0.1",
		}

		ioutil.WriteFile(filepath.Join(tmpDir, "secret"), []byte("some-bosh-secret"), 0600)

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
				mockUI.EXPECT().Say(fmt.Sprintf(
					`unset BOSH_SOME_VAR;
unset BOSH_SOME_OTHER_VAR;
export BOSH_ENVIRONMENT="10.0.0.1";
export BOSH_CLIENT="admin";
export BOSH_CLIENT_SECRET="some-bosh-secret";
export BOSH_CA_CERT="%s";
export BOSH_GW_HOST="10.0.0.1";
export BOSH_GW_PRIVATE_KEY="%s";
export BOSH_GW_USER="jumpbox";`,
					filepath.Join(tmpDir, "ca.crt"),
					filepath.Join(tmpDir, "jumpbox.key"),
				))

				Expect(boshCmd.Env()).To(Succeed())
			})
		})
	})
})
