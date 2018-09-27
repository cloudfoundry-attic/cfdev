package os_test

import (
	"code.cloudfoundry.org/cfdev/analyticsd/os"
	"code.cloudfoundry.org/cfdev/analyticsd/os/mocks"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("OS version", func() {

	var (
		obj            *os.OS
		mockController *gomock.Controller
		mockRunner     *mocks.MockRunner
	)

	BeforeEach(func() {
		mockController = gomock.NewController(GinkgoT())
		mockRunner = mocks.NewMockRunner(mockController)

		obj = &os.OS{
			Runner: mockRunner,
		}
	})

	AfterEach(func() {
		mockController.Finish()
	})

	Context("when a valid version is returned", func() {
		It("returns the os version", func() {
			mockRunner.EXPECT().Output("sw_vers").Return([]byte(`
ProductName:	Mac OS X
ProductVersion:	10.13.6
BuildVersion:	17G65
		`), nil)

			Expect(obj.Version()).To(Equal("10.13.6"))
		})
	})

	Context("when an invalid version is returned", func() {
		It("returns a useful error message", func() {
			mockRunner.EXPECT().Output("sw_vers").Return([]byte(`
LeProductName:	Mac OS X
LeProducteVersion:	10.13.6
LaBuildVersion:	17G65
		`), nil)

			_, err := obj.Version()
			Expect(err.Error()).To(ContainSubstring("failed to parse os version out of:"))
		})
	})
})
