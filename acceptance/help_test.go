package acceptance

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"os/exec"

	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
)

var _ = Describe("help", func() {
	It("running 'cf dev' provides help", func() {
		cmd := exec.Command("cf", "dev")
		session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
		Expect(err).ToNot(HaveOccurred())

		Eventually(session).Should(gexec.Exit(0))
		Expect(session).To(gbytes.Say("Usage:"))
		Expect(session).To(gbytes.Say("Available Commands:"))
	})

	It("running 'cf dev help' provides help", func() {
		cmd := exec.Command("cf", "dev", "help")
		session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
		Expect(err).ToNot(HaveOccurred())

		Eventually(session).Should(gexec.Exit(0))
		Expect(session).To(gbytes.Say("Usage:"))
		Expect(session).To(gbytes.Say("Available Commands:"))
	})
})
