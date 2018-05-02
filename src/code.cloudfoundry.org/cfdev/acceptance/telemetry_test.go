package acceptance

import (
	"io/ioutil"
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"os/exec"

	"fmt"

	"path"

	"time"

	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
)

var _ = Describe("hyperkit starts and telemetry", func() {
	var (
		stateDir string
		cacheDir string
		session  *gexec.Session
		err      error
	)

	BeforeEach(func() {
		cacheDir = filepath.Join(cfdevHome, "cache")
		stateDir = filepath.Join(cfdevHome, "state")

		if os.Getenv("CFDEV_PLUGIN_PATH") == "" {
			SetupDependencies(cacheDir)
			os.Setenv("CFDEV_SKIP_ASSET_CHECK", "true")
		}

		os.RemoveAll(path.Join(cfdevHome, "analytics"))
	})

	AfterEach(func() {
		session.Kill()
	})

	It("optout", func() {
		cmd := exec.Command("/usr/local/bin/cf", "dev", "start")
		inWriter, _ := cmd.StdinPipe()
		session, err = gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
		Expect(err).ToNot(HaveOccurred())

		Eventually(session).Should(gbytes.Say("Are you ok with CF Dev periodically capturing anonymized telemetry"))

		fmt.Fprintln(inWriter, "no")

		Eventually(func() (string, error) {
			contents, err := ioutil.ReadFile(filepath.Join(cfdevHome, "analytics", "analytics.txt"))
			return string(contents), err
		}).Should(Equal("optout"))
	})

	It("optin", func() {
		cmd := exec.Command("/usr/local/bin/cf", "dev", "start")
		inWriter, _ := cmd.StdinPipe()
		session, err = gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
		Expect(err).ToNot(HaveOccurred())

		Eventually(session).Should(gbytes.Say("Are you ok with CF Dev periodically capturing anonymized telemetry"))

		fmt.Fprintln(inWriter, "yes")

		Eventually(func() (string, error) {
			contents, err := ioutil.ReadFile(filepath.Join(cfdevHome, "analytics", "analytics.txt"))
			return string(contents), err
		}).Should(Equal("optin"))
	})

	It("is already opted in", func() {
		err := os.MkdirAll(path.Join(cfdevHome, "analytics"), 0755)
		Expect(err).ToNot(HaveOccurred())
		err = ioutil.WriteFile(path.Join(cfdevHome, "analytics", "analytics.txt"), []byte("optin"), 0755)
		Expect(err).ToNot(HaveOccurred())

		cmd := exec.Command("/usr/local/bin/cf", "dev", "start")
		session, err = gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
		Expect(err).ToNot(HaveOccurred())

		Consistently(session, time.Second).ShouldNot(gbytes.Say("Are you ok with CF Dev periodically capturing anonymized telemetry"))

		Expect(ioutil.ReadFile(filepath.Join(cfdevHome, "analytics", "analytics.txt"))).Should(Equal([]byte("optin")))
	})
})
