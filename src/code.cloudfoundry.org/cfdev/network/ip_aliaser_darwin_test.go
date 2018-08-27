package network_test

import (
	//"os/exec"

	"code.cloudfoundry.org/cfdev/network"
	"code.cloudfoundry.org/cfdev/network/mocks"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("IP Aliaser - Darwin", func() {
	var (
		hostnet          *network.HostNet
		mockCfdevdClient *mocks.MockCfdevdClient
		mockController   *gomock.Controller
	)

	BeforeEach(func() {
		mockController = gomock.NewController(GinkgoT())
		mockCfdevdClient = mocks.NewMockCfdevdClient(mockController)

		hostnet = &network.HostNet{
			CfdevdClient: mockCfdevdClient,
		}
	})

	AfterEach(func() {
		mockController.Finish()
	})

	Describe("AddLoopbackAliases", func() {
		It("calls cfdevd.AddLoopbackAliases", func() {
			mockCfdevdClient.EXPECT().AddIPAlias()
			Expect(hostnet.AddLoopbackAliases()).To(Succeed())
		})
	})

	Describe("RemoveLoopbackAliases", func() {
		It("calls cfdevd.RemoveLoopbackAliases", func() {
			mockCfdevdClient.EXPECT().RemoveIPAlias()
			Expect(hostnet.RemoveLoopbackAliases()).To(Succeed())
		})
	})
})
