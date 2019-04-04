package provision_test

import (
	"code.cloudfoundry.org/cfdev/cmd/provision"
	"code.cloudfoundry.org/cfdev/cmd/provision/mocks"
	"code.cloudfoundry.org/cfdev/cmd/start"
	"code.cloudfoundry.org/cfdev/config"
	"code.cloudfoundry.org/cfdev/workspace"
	"code.cloudfoundry.org/cli/cf/errors"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
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
				StateDir: "some-state-dir",
			},
		}
	})

	Describe("happy path", func() {
		It("deploys bosh and services", func() {
			gomock.InOrder(
				mockMetadataReader.EXPECT().Metadata().Return(workspace.Metadata{
					Version: "v5",
				}, nil),
				mockProvisioner.EXPECT().Ping(gomock.Any()),
				mockUI.EXPECT().Say("Deploying the BOSH Director..."),
				mockProvisioner.EXPECT().DeployBosh(),
				mockProvisioner.EXPECT().WhiteListServices("", nil).Return([]workspace.Service{}, nil),
				mockProvisioner.EXPECT().DeployServices(mockUI, []workspace.Service{}, nil),
			)

			err := cmd.Execute(start.Args{})
			Expect(err).NotTo(HaveOccurred())
		})
	})

	Describe("when version is not compatible", func() {
		It("return an error", func() {
			gomock.InOrder(
				mockMetadataReader.EXPECT().Metadata().Return(workspace.Metadata{
					Version: "v10",
				}, nil),
			)

			err := cmd.Execute(start.Args{})
			Expect(err).To(MatchError(ContainSubstring("version is incompatible")))
		})
	})

	Describe("when docker registry flags are present", func() {
		It("pass them to services", func() {
			gomock.InOrder(
				mockMetadataReader.EXPECT().Metadata().Return(workspace.Metadata{
					Version: "v5",
				}, nil),
				mockProvisioner.EXPECT().Ping(gomock.Any()),
				mockUI.EXPECT().Say("Deploying the BOSH Director..."),
				mockProvisioner.EXPECT().DeployBosh(),
				mockProvisioner.EXPECT().WhiteListServices("", nil).Return([]workspace.Service{}, nil),
				mockProvisioner.EXPECT().DeployServices(mockUI, []workspace.Service{}, []string{"domain1.com", "domain2.com"}),
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
				mockMetadataReader.EXPECT().Metadata().Return(workspace.Metadata{
					Version: "v5",
				}, nil),
				mockProvisioner.EXPECT().Ping(gomock.Any()).Return(errors.New("not running")),
			)

			err := cmd.Execute(start.Args{})
			Expect(err).To(MatchError(ContainSubstring("VM is not running")))
		})
	})

	AfterEach(func() {
		mockController.Finish()
	})
})
