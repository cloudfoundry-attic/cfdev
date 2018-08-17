package bosh_test

import (
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
		mockController  *gomock.Controller
		mockProvisioner *mocks.MockProvisioner
		mockUI          *mocks.MockUI
		tmpDir          string
		boshCmd         *cmd.Bosh
	)

	BeforeEach(func() {
		mockController = gomock.NewController(GinkgoT())
		mockProvisioner = mocks.NewMockProvisioner(mockController)
		mockUI = mocks.NewMockUI(mockController)

		var err error
		tmpDir, err = ioutil.TempDir("", "cmd-bosh-test")
		Expect(err).NotTo(HaveOccurred())
		boshCmd = &cmd.Bosh{
			Provisioner: mockProvisioner,
			StateDir:    tmpDir,
			UI:          mockUI,
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
					CACertificate:     "some-ca-cert",
					DirectorAddress:   "some-director-address",
					GatewayHost:       "some-gateway-host",
					GatewayPrivateKey: "some-gateway-private-key",
					GatewayUsername:   "some-gateway-username",
				}, nil)
				mockUI.EXPECT().Say(fmt.Sprintf(
					`export BOSH_ENVIRONMENT="some-director-address";
export BOSH_CLIENT="some-admin-username";
export BOSH_CLIENT_SECRET="some-admin-password";
export BOSH_CA_CERT="%s";
export BOSH_GW_HOST="some-gateway-host";
export BOSH_GW_PRIVATE_KEY="%s";
export BOSH_GW_USER="some-gateway-username";`,
					filepath.Join(tmpDir, "bosh-ca-cert"),
					filepath.Join(tmpDir, "bosh-gw-key"),
				),
				)
				Expect(boshCmd.Env()).To(Succeed())

				contents, err := ioutil.ReadFile(filepath.Join(tmpDir, "bosh-ca-cert"))
				Expect(err).NotTo(HaveOccurred())
				Expect(string(contents)).To(Equal("some-ca-cert"))

				contents, err = ioutil.ReadFile(filepath.Join(tmpDir, "bosh-gw-key"))
				Expect(err).NotTo(HaveOccurred())
				Expect(string(contents)).To(Equal("some-gateway-private-key"))
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
					CACertificate:     "some-ca-cert",
					DirectorAddress:   "some-director-address",
					GatewayHost:       "some-gateway-host",
					GatewayPrivateKey: "some-gateway-private-key",
					GatewayUsername:   "some-gateway-username",
				}, nil)
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
					filepath.Join(tmpDir, "bosh-ca-cert"),
					filepath.Join(tmpDir, "bosh-gw-key"),
				),
				)
				Expect(boshCmd.Env()).To(Succeed())
				contents, err := ioutil.ReadFile(filepath.Join(tmpDir, "bosh-ca-cert"))
				Expect(err).NotTo(HaveOccurred())
				Expect(string(contents)).To(Equal("some-ca-cert"))

				contents, err = ioutil.ReadFile(filepath.Join(tmpDir, "bosh-gw-key"))
				Expect(err).NotTo(HaveOccurred())
				Expect(string(contents)).To(Equal("some-gateway-private-key"))
			})
		})
	})
})
