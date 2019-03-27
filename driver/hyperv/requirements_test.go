package hyperv_test

import (
	"code.cloudfoundry.org/cfdev/driver/hyperv"
	"code.cloudfoundry.org/cfdev/errors"
	"code.cloudfoundry.org/cfdev/host/mocks"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Host", func() {

	var (
		mockController *gomock.Controller
		mockPowershell *mocks.MockPowershell
		d *hyperv.HyperV

		adminQueryStr = `(New-Object Security.Principal.WindowsPrincipal([Security.Principal.WindowsIdentity]::GetCurrent())).IsInRole([Security.Principal.WindowsBuiltInRole]::Administrator)`
	)

	BeforeEach(func() {
		mockController = gomock.NewController(GinkgoT())
		mockPowershell = mocks.NewMockPowershell(mockController)

		d = &hyperv.HyperV{
			Powershell:    mockPowershell,
		}
	})

	AfterEach(func() {
		mockController.Finish()
	})

	Describe("check requirements", func() {
		Context("when not running in an admin shell", func() {
			It("returns an error", func() {
				mockPowershell.EXPECT().Output(adminQueryStr).Return("False", nil)

				err := d.CheckRequirements()
				Expect(err.Error()).To(ContainSubstring(`Running without admin privileges: You must run cf dev with an admin privileged powershell`))
				Expect(errors.SafeError(err)).To(Equal("Running without admin privileges"))
			})
		})

		Context("when running in an admin shell", func() {
			Context("Hyper-V is enabled on a Windows 10 machine", func() {
				It("succeeds", func() {
					gomock.InOrder(
						mockPowershell.EXPECT().Output(adminQueryStr).Return("True", nil),
						mockPowershell.EXPECT().Output(`(Get-WindowsOptionalFeature -FeatureName Microsoft-Hyper-V -Online).State`).Return("Enabled", nil),
						mockPowershell.EXPECT().Output(`(Get-WindowsOptionalFeature -FeatureName Microsoft-Hyper-V-Management-PowerShell -Online).State`).Return("Enabled", nil),
					)

					Expect(d.CheckRequirements()).To(Succeed())
				})
			})

			Context("Microsoft-Hyper-V is disabled", func() {
				It("returns an error", func() {
					gomock.InOrder(
						mockPowershell.EXPECT().Output(adminQueryStr).Return("True", nil),
						mockPowershell.EXPECT().Output(`(Get-WindowsOptionalFeature -FeatureName Microsoft-Hyper-V -Online).State`).Return("Disabled", nil),
					)

					err := d.CheckRequirements()
					Expect(err.Error()).To(ContainSubstring(`Microsoft-Hyper-V disabled: You must first enable Hyper-V on your machine`))
					Expect(errors.SafeError(err)).To(Equal("Microsoft-Hyper-V disabled"))
				})
			})

			Context("Microsoft-Hyper-V-Management-PowerShell is disabled", func() {
				It("returns an error", func() {
					gomock.InOrder(
						mockPowershell.EXPECT().Output(adminQueryStr).Return("True", nil),
						mockPowershell.EXPECT().Output(`(Get-WindowsOptionalFeature -FeatureName Microsoft-Hyper-V -Online).State`).Return("Enabled", nil),
						mockPowershell.EXPECT().Output(`(Get-WindowsOptionalFeature -FeatureName Microsoft-Hyper-V-Management-PowerShell -Online).State`).Return("Disabled", nil),
					)

					err := d.CheckRequirements()
					Expect(err.Error()).To(ContainSubstring(`Microsoft-Hyper-V-Management-PowerShell disabled: You must first enable Hyper-V on your machine`))
					Expect(errors.SafeError(err)).To(Equal("Microsoft-Hyper-V-Management-PowerShell disabled"))
				})
			})
		})
	})
})

