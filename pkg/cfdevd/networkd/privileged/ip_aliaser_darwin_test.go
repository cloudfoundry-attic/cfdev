package privileged

import (
	"os/exec"

	"code.cloudfoundry.org/cfdev/pkg/cfdevd/networkd"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
)

var _ = Describe("IP Aliaser - Darwin", func() {
	var hostnet *networkd.HostNetD

	BeforeEach(func() {
		hostnet = &networkd.HostNetD{}
	})

	Describe("AddLoopbackAliases", func() {
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
			Expect(hostnet.AddLoopbackAliases("123.123.123.123", "6.6.6.6")).To(Succeed())
			session, err := gexec.Start(exec.Command("ifconfig", "lo0"), GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())
			Eventually(session).Should(gexec.Exit(0))
			Expect(string(session.Out.Contents())).To(ContainSubstring("inet 123.123.123.123 netmask 0xffffffff"))
			Expect(string(session.Out.Contents())).To(ContainSubstring("inet 6.6.6.6 netmask 0xffffffff"))
		})
	})

	Describe("RemoveLoopbackAliases", func() {
		Context("when the aliases exists", func() {
			BeforeEach(func() {
				session, err := gexec.Start(
					exec.Command("sudo", "-S", "ifconfig", "lo0", "add", "123.123.123.123/32"),
					GinkgoWriter,
					GinkgoWriter,
				)
				Expect(err).NotTo(HaveOccurred())
				Eventually(session).Should(gexec.Exit(0))

				session, err = gexec.Start(
					exec.Command("sudo", "-S", "ifconfig", "lo0", "add", "6.6.6.6/32"),
					GinkgoWriter,
					GinkgoWriter,
				)
				Expect(err).NotTo(HaveOccurred())
				Eventually(session).Should(gexec.Exit(0))

				session, err = gexec.Start(
					exec.Command("ifconfig", "lo0"),
					GinkgoWriter,
					GinkgoWriter,
				)

				Expect(err).NotTo(HaveOccurred())
				Eventually(session).Should(gexec.Exit(0))
				Expect(string(session.Out.Contents())).To(ContainSubstring("inet 123.123.123.123 netmask 0xffffffff"))
				Expect(string(session.Out.Contents())).To(ContainSubstring("inet 6.6.6.6 netmask 0xffffffff"))
			})

			It("removes aliases from the lo0 interface", func() {
				Expect(hostnet.RemoveLoopbackAliases("123.123.123.123", "6.6.6.6")).To(Succeed())
				session, err := gexec.Start(exec.Command("ifconfig", "lo0"), GinkgoWriter, GinkgoWriter)
				Expect(err).NotTo(HaveOccurred())
				Eventually(session).Should(gexec.Exit(0))
				Expect(string(session.Out.Contents())).NotTo(ContainSubstring("123.123.123.123"))
				Expect(string(session.Out.Contents())).NotTo(ContainSubstring("6.6.6.6"))
			})
		})

		Context("when the aliases do not exist", func() {
			BeforeEach(func() {
				session, err := gexec.Start(exec.Command("ifconfig", "lo0"), GinkgoWriter, GinkgoWriter)
				Expect(err).NotTo(HaveOccurred())
				Eventually(session).Should(gexec.Exit(0))
				Expect(string(session.Out.Contents())).NotTo(ContainSubstring("123.123.123.123"))
				Expect(string(session.Out.Contents())).NotTo(ContainSubstring("6.6.6.6"))
			})

			It("succeeds", func() {
				Expect(hostnet.RemoveLoopbackAliases("123.123.123.123", "6.6.6.6")).To(Succeed())
			})
		})
	})
})
