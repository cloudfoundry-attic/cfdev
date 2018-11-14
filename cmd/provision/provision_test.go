package provision_test

import (
	"code.cloudfoundry.org/cfdev/cmd/provision"
	"code.cloudfoundry.org/cfdev/cmd/provision/mocks"
	"code.cloudfoundry.org/cfdev/cmd/start"
	"code.cloudfoundry.org/cfdev/config"
	"code.cloudfoundry.org/cfdev/metadata"
	prvsion "code.cloudfoundry.org/cfdev/provision"
	"code.cloudfoundry.org/cli/cf/errors"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	"path/filepath"

	. "github.com/onsi/gomega"
)

var _ = Describe("Provision", func() {
	var (
		mockController     *gomock.Controller
		mockUI             *mocks.MockUI
		mockMetadataReader *mocks.MockMetaDataReader
		mockProvisioner    *mocks.MockProvisioner
		cmd                *provision.Provision
	)

	BeforeEach(func() {
		mockController = gomock.NewController(GinkgoT())
		mockUI = mocks.NewMockUI(mockController)
		mockProvisioner = mocks.NewMockProvisioner(mockController)
		mockMetadataReader = mocks.NewMockMetaDataReader(mockController)

		localExitChan := make(chan struct{}, 3)

		cmd = &provision.Provision{
			Exit:           localExitChan,
			UI:             mockUI,
			Provisioner:    mockProvisioner,
			MetaDataReader: mockMetadataReader,
			Config: config.Config{
				CacheDir: "some-cache-dir",
			},
		}
	})

	Describe("happy path", func() {
		It("deploys bosh and cf and services", func() {
			gomock.InOrder(
				mockMetadataReader.EXPECT().Read(filepath.Join("some-cache-dir", "metadata.yml")).Return(metadata.Metadata{
					Version: "v3",
				}, nil),
				mockProvisioner.EXPECT().Ping(),
				mockUI.EXPECT().Say("Deploying the BOSH Director..."),
				mockProvisioner.EXPECT().DeployBosh(),
				mockUI.EXPECT().Say("Deploying CF..."),
				mockProvisioner.EXPECT().DeployCloudFoundry(mockUI, nil),
				mockProvisioner.EXPECT().WhiteListServices("", nil).Return([]prvsion.Service{}, nil),
				mockProvisioner.EXPECT().DeployServices(mockUI, []prvsion.Service{}),
			)

			err := cmd.Execute(start.Args{})
			Expect(err).NotTo(HaveOccurred())
		})
	})

	Describe("when version is not compatible", func() {
		It("return an error", func() {
			gomock.InOrder(
				mockMetadataReader.EXPECT().Read(filepath.Join("some-cache-dir", "metadata.yml")).Return(metadata.Metadata{
					Version: "v10",
				}, nil),
			)

			err := cmd.Execute(start.Args{})
			Expect(err).To(MatchError(ContainSubstring("version is incompatible")))
		})
	})

	Describe("when docker registry flags are present", func() {
		It("pass them to CF", func() {
			gomock.InOrder(
				mockMetadataReader.EXPECT().Read(filepath.Join("some-cache-dir", "metadata.yml")).Return(metadata.Metadata{
					Version: "v3",
				}, nil),
				mockProvisioner.EXPECT().Ping(),
				mockUI.EXPECT().Say("Deploying the BOSH Director..."),
				mockProvisioner.EXPECT().DeployBosh(),
				mockUI.EXPECT().Say("Deploying CF..."),
				mockProvisioner.EXPECT().DeployCloudFoundry(mockUI, []string{"domain1.com", "domain2.com"}),
				mockProvisioner.EXPECT().WhiteListServices("", nil).Return([]prvsion.Service{}, nil),
				mockProvisioner.EXPECT().DeployServices(mockUI, []prvsion.Service{}),
			)

			err := cmd.Execute(start.Args{
				Registries: "domain1.com,domain2.com",
			})
			Expect(err).NotTo(HaveOccurred())
		})
	})

	Describe("when the vm is not running", func() {
		It("return an error", func() {
			gomock.InOrder(
				mockMetadataReader.EXPECT().Read(filepath.Join("some-cache-dir", "metadata.yml")).Return(metadata.Metadata{
					Version: "v3",
				}, nil),
				mockProvisioner.EXPECT().Ping().Return(errors.New("not running")),
			)

			err := cmd.Execute(start.Args{})
			Expect(err).To(MatchError(ContainSubstring("VM is not running")))
		})
	})

	AfterEach(func() {
		mockController.Finish()
	})
})
