package acceptance

import (
	"os"
	"os/exec"
	"path/filepath"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
)

var _ = Describe("help", func() {
	BeforeEach(func() {
		if os.Getenv("CFDEV_PLUGIN_PATH") == "" {
			SetupDependencies(filepath.Join(cfdevHome, "cache"))
			os.Setenv("CFDEV_SKIP_ASSET_CHECK", "true")
		}
	})

	AfterEach(func() {
		gexec.Kill()
		os.Unsetenv("CFDEV_SKIP_ASSET_CHECK")
	})

	It("running 'cf dev' provides help", func() {
		cmd := exec.Command("cf", "dev")
		session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
		Expect(err).ToNot(HaveOccurred())

		Eventually(session, 10*time.Second).Should(gexec.Exit(0))
		Expect(session).To(gbytes.Say("Usage:"))
		Expect(session).To(gbytes.Say("Available Commands:"))
	})

	It("running 'cf dev help' provides help", func() {
		cmd := exec.Command("cf", "dev", "help")
		session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
		Expect(err).ToNot(HaveOccurred())

		Eventually(session, 10*time.Second).Should(gexec.Exit(0))
		Expect(session).To(gbytes.Say("Usage:"))
		Expect(session).To(gbytes.Say("Available Commands:"))
	})
})
