package privileged_test

import (
	"os/exec"

	"code.cloudfoundry.org/cfdev/network"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
)

var _ = Describe("Start", func() {
	Describe("AddLoopbackAlias", func() {
		BeforeEach(func() {
			hostNet := network.HostNet{}
			Expect(hostNet.AddLoopbackAliases("123.123.123.123", "6.6.6.6")).To(Succeed())
		})

		AfterEach(func() {
			session, err := gexec.Start(
				exec.Command("sudo", "ifconfig", "lo0", "inet", "123.123.123.123/32", "remove"),
				GinkgoWriter,
				GinkgoWriter,
			)
			Expect(err).NotTo(HaveOccurred())
			Eventually(session).Should(gexec.Exit(0))

			session, err = gexec.Start(
				exec.Command("sudo", "ifconfig", "lo0", "inet", "6.6.6.6/32", "remove"),
				GinkgoWriter,
				GinkgoWriter,
			)
			Expect(err).NotTo(HaveOccurred())
			Eventually(session).Should(gexec.Exit(0))
		})

		It("adds aliases to the lo0 interface", func() {
			session, err := gexec.Start(exec.Command("ifconfig", "lo0"), GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())
			Eventually(session).Should(gexec.Exit(0))
			Expect(session.Out.Contents()).To(ContainSubstring("inet 123.123.123.123 netmask 0xffffffff"))
			Expect(session.Out.Contents()).To(ContainSubstring("inet 6.6.6.6 netmask 0xffffffff"))
		})
	})
})
