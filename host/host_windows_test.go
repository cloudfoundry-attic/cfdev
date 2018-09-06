package host_test

import (
	"os/exec"

	"code.cloudfoundry.org/cfdev/errors"
	"code.cloudfoundry.org/cfdev/host"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Host", func() {
	Describe("check requirements", func() {
		Context("when running in an admin shell", func() {
			//we always run windows tests in an admin shell
			Context("when hyperv is enabled", func() {
				// we assume tests always run on a machine with hyperv enabled
				It("succeeds", func() {
					h := &host.Host{}
					Expect(h.CheckRequirements()).To(Succeed())
				})
			})
			Context("when hyperv is disabled", func() {
				BeforeEach(func() {
					cmd := exec.Command("powershell.exe", "-Command",
						"Disable-WindowsOptionalFeature -Online -FeatureName Microsoft-Hyper-V-All -NoRestart")
					output, err := cmd.CombinedOutput()
					Expect(err).NotTo(HaveOccurred(), string(output))
				})

				AfterEach(func() {
					cmd := exec.Command("powershell.exe", "-Command",
						"Enable-WindowsOptionalFeature -Online -FeatureName Microsoft-Hyper-V-All -NoRestart")
					output, err := cmd.CombinedOutput()
					Expect(err).NotTo(HaveOccurred(), string(output))
				})

				It("fails", func() {
					h := &host.Host{}
					err := h.CheckRequirements()
					Expect(err.Error()).To(ContainSubstring(`Hyper-V disabled: You must first enable Hyper-V on your machine`))
					Expect(errors.SafeError(err)).To(Equal("Hyper-V disabled"))
				})
			})
		})
	})
})
