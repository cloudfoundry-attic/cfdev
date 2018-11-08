package bosh_test

import (
	"code.cloudfoundry.org/cfdev/cfanalytics"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"code.cloudfoundry.org/cfdev/bosh"
	cmd "code.cloudfoundry.org/cfdev/cmd/bosh"
	"code.cloudfoundry.org/cfdev/cmd/bosh/mocks"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Bosh", func() {
	var (
		mockController      *gomock.Controller
		mockProvisioner     *mocks.MockProvisioner
		mockUI              *mocks.MockUI
		tmpDir              string
		boshCmd             *cmd.Bosh
		mockAnalyticsClient *mocks.MockAnalyticsClient
	)

	BeforeEach(func() {
		mockController = gomock.NewController(GinkgoT())
		mockProvisioner = mocks.NewMockProvisioner(mockController)
		mockAnalyticsClient = mocks.NewMockAnalyticsClient(mockController)
		mockUI = mocks.NewMockUI(mockController)

		var err error
		tmpDir, err = ioutil.TempDir("", "cmd-bosh-test")
		Expect(err).NotTo(HaveOccurred())
		boshCmd = &cmd.Bosh{
			Provisioner: mockProvisioner,
			StateDir:    tmpDir,
			UI:          mockUI,
			Analytics:   mockAnalyticsClient,
		}
	})

	AfterEach(func() {
		mockController.Finish()
		os.RemoveAll(tmpDir)
	})

	Describe("Env", func() {
		Context("when no BOSH_* env vars are currently set", func() {
			BeforeEach(func() {
				for _, envvar := range os.Environ() {
					if strings.HasPrefix(envvar, "BOSH_") {
						key := strings.Split(envvar, "=")[0]
						os.Unsetenv(key)
					}
				}
			})

			It("print the export statements", func() {
				mockProvisioner.EXPECT().FetchBOSHConfig().Return(bosh.Config{
					AdminUsername:     "some-admin-username",
					AdminPassword:     "some-admin-password",
					CACertificate:     filepath.Join(tmpDir, "some-ca-cert"),
					DirectorAddress:   "some-director-address",
					GatewayHost:       "some-gateway-host",
					GatewayPrivateKey: filepath.Join(tmpDir, "some-gateway-private-key"),
					GatewayUsername:   "some-gateway-username",
				}, nil)

				mockAnalyticsClient.EXPECT().Event(cfanalytics.BOSH_ENV)

				mockUI.EXPECT().Say(fmt.Sprintf(
					`export BOSH_ENVIRONMENT="some-director-address";
export BOSH_CLIENT="some-admin-username";
export BOSH_CLIENT_SECRET="some-admin-password";
export BOSH_CA_CERT="%s";
export BOSH_GW_HOST="some-gateway-host";
export BOSH_GW_PRIVATE_KEY="%s";
export BOSH_GW_USER="some-gateway-username";`,
					filepath.Join(tmpDir, "some-ca-cert"),
					filepath.Join(tmpDir, "some-gateway-private-key"),
				),
				)
				Expect(boshCmd.Env()).To(Succeed())
			})
		})

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
				mockProvisioner.EXPECT().FetchBOSHConfig().Return(bosh.Config{
					AdminUsername:     "some-admin-username",
					AdminPassword:     "some-admin-password",
					CACertificate:     filepath.Join(tmpDir, "some-ca-cert"),
					DirectorAddress:   "some-director-address",
					GatewayHost:       "some-gateway-host",
					GatewayPrivateKey: filepath.Join(tmpDir, "some-gateway-private-key"),
					GatewayUsername:   "some-gateway-username",
				}, nil)
				mockAnalyticsClient.EXPECT().Event(cfanalytics.BOSH_ENV)
				mockUI.EXPECT().Say(fmt.Sprintf(
					`unset BOSH_SOME_VAR;
unset BOSH_SOME_OTHER_VAR;
export BOSH_ENVIRONMENT="some-director-address";
export BOSH_CLIENT="some-admin-username";
export BOSH_CLIENT_SECRET="some-admin-password";
export BOSH_CA_CERT="%s";
export BOSH_GW_HOST="some-gateway-host";
export BOSH_GW_PRIVATE_KEY="%s";
export BOSH_GW_USER="some-gateway-username";`,
					filepath.Join(tmpDir, "some-ca-cert"),
					filepath.Join(tmpDir, "some-gateway-private-key"),
				),
				)
				Expect(boshCmd.Env()).To(Succeed())
			})
		})
	})
})
