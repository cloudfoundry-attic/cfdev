package deploy_service_test

import (
	"code.cloudfoundry.org/cfdev/cmd/deploy-service"
	"code.cloudfoundry.org/cfdev/cmd/deploy-service/mocks"
	"code.cloudfoundry.org/cfdev/config"
	"code.cloudfoundry.org/cfdev/workspace"
	"errors"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("DeployService", func() {
	var (
		mockController     *gomock.Controller
		mockMetadataReader *mocks.MockMetaDataReader
		mockProvisioner    *mocks.MockProvisioner
		mockUI             *mocks.MockUI
		mockAnalytics      *mocks.MockAnalytics
		cmd                *deploy_service.DeployService
	)
	BeforeEach(func() {
		mockController = gomock.NewController(GinkgoT())
		mockUI = mocks.NewMockUI(mockController)
		mockMetadataReader = mocks.NewMockMetaDataReader(mockController)
		mockProvisioner = mocks.NewMockProvisioner(mockController)
		mockAnalytics = mocks.NewMockAnalytics(mockController)

		cmd = &deploy_service.DeployService{
			UI:             mockUI,
			MetaDataReader: mockMetadataReader,
			Provisioner:    mockProvisioner,
			Config: config.Config{
				StateDir: "some-state-dir",
			},
			Analytics: mockAnalytics,
		}
	})

	AfterEach(func() {})

	Describe("happy path", func() {
		It("deploys a new service", func() {
			service := workspace.Service{
				Name: "some-service",
			}
			mockMetadataReader.EXPECT().Metadata().Return(workspace.Metadata{
				Version:  "v5",
				Services: []workspace.Service{service},
			}, nil)

			mockProvisioner.EXPECT().Ping(gomock.Any()).Return(nil)
			mockProvisioner.EXPECT().GetWhiteListedService("some-service", []workspace.Service{service}).Return(&service, nil)
			mockProvisioner.EXPECT().DeployServices(mockUI, []workspace.Service{service}, []string{}).Return(nil)

			mockAnalytics.EXPECT().Event("deployed service", map[string]interface{}{"name": "some-service"})

			err := cmd.Execute(deploy_service.Args{
				Service: "some-service",
			})
			Expect(err).NotTo(HaveOccurred())
		})
	})

	Describe("When cf dev is not running", func() {
		It("returns an error", func() {
			service := workspace.Service{
				Name: "some-service",
			}
			mockMetadataReader.EXPECT().Metadata().Return(workspace.Metadata{
				Version:  "v5",
				Services: []workspace.Service{service},
			}, nil)
			mockProvisioner.EXPECT().Ping(gomock.Any()).Return(errors.New("some issue happened"))

			err := cmd.Execute(deploy_service.Args{
				Service: "some-service",
			})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("cf dev is not running"))
		})
	})

	Describe("When service is not whitelisted", func() {
		It("returns an error", func() {
			service := workspace.Service{
				Name: "some-service",
			}
			mockMetadataReader.EXPECT().Metadata().Return(workspace.Metadata{
				Version:  "v5",
				Services: []workspace.Service{service},
			}, nil)
			mockProvisioner.EXPECT().Ping(gomock.Any()).Return(nil)
			mockProvisioner.EXPECT().GetWhiteListedService(
				"some-service",
				[]workspace.Service{service}).Return(&service, errors.New("unknown service"))

			err := cmd.Execute(deploy_service.Args{
				Service: "some-service",
			})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("Failed to whitelist service"))
		})
	})
})
