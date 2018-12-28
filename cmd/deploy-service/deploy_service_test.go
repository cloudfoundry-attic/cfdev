package deploy_service_test

import (
	"code.cloudfoundry.org/cfdev/cmd/deploy-service"
	"code.cloudfoundry.org/cfdev/cmd/deploy-service/mocks"
	"code.cloudfoundry.org/cfdev/config"
	"code.cloudfoundry.org/cfdev/metadata"
	"code.cloudfoundry.org/cfdev/provision"
	"errors"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"path/filepath"
)

var _ = Describe("DeployService", func() {
	var (
		mockController     *gomock.Controller
		mockMetadataReader *mocks.MockMetaDataReader
		mockProvisioner    *mocks.MockProvisioner
		mockUI             *mocks.MockUI
		cmd                *deploy_service.DeployService
	)
	BeforeEach(func() {
		mockController = gomock.NewController(GinkgoT())
		mockUI = mocks.NewMockUI(mockController)
		mockMetadataReader = mocks.NewMockMetaDataReader(mockController)
		mockProvisioner = mocks.NewMockProvisioner(mockController)

		cmd = &deploy_service.DeployService{
			UI:             mockUI,
			MetaDataReader: mockMetadataReader,
			Provisioner:    mockProvisioner,
			Config: config.Config{
				CacheDir: "some-cache-dir",
			},
		}
	})

	AfterEach(func() {})

	Describe("happy path", func() {
		It("deploys a new service", func() {
			service := provision.Service{
				Name: "some-service",
			}
			mockMetadataReader.EXPECT().Read(filepath.Join("some-cache-dir", "metadata.yml")).Return(metadata.Metadata{
				Version:  "v3",
				Services: []provision.Service{service},
			}, nil)
			mockProvisioner.EXPECT().Ping().Return(nil)
			mockProvisioner.EXPECT().GetWhiteListedService("some-service", []provision.Service{service}).Return(&service, nil)

			mockProvisioner.EXPECT().DeployServices(mockUI, []provision.Service{service}).Return(nil)

			err := cmd.Execute(deploy_service.Args{
				Service: "some-service",
			})
			Expect(err).NotTo(HaveOccurred())
		})
	})

	Describe("When cf dev is not running", func() {
		It("returns an error", func() {
			service := provision.Service{
				Name: "some-service",
			}
			mockMetadataReader.EXPECT().Read(filepath.Join("some-cache-dir", "metadata.yml")).Return(metadata.Metadata{
				Version:  "v3",
				Services: []provision.Service{service},
			}, nil)
			mockProvisioner.EXPECT().Ping().Return(errors.New("some issue happened"))

			err := cmd.Execute(deploy_service.Args{
				Service: "some-service",
			})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("cf dev is not running"))
		})
	})

	Describe("When service is not whitelisted", func() {
		It("returns an error", func() {
			service := provision.Service{
				Name: "some-service",
			}
			mockMetadataReader.EXPECT().Read(filepath.Join("some-cache-dir", "metadata.yml")).Return(metadata.Metadata{
				Version:  "v3",
				Services: []provision.Service{service},
			}, nil)
			mockProvisioner.EXPECT().Ping().Return(nil)
			mockProvisioner.EXPECT().GetWhiteListedService(
				"some-service",
				[]provision.Service{service}).Return(&service, errors.New("unknown service"))


			err := cmd.Execute(deploy_service.Args{
				Service: "some-service",
			})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("Failed to whitelist service"))
		})
	})
})
