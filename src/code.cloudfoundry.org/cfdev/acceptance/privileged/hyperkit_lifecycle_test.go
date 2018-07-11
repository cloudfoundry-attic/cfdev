package privileged_test

import (
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"

	"time"

	"io"
	"io/ioutil"
	"net/http"

	. "code.cloudfoundry.org/cfdev/acceptance"
	"code.cloudfoundry.org/garden/client"
	"code.cloudfoundry.org/garden/client/connection"
	"github.com/cloudfoundry-incubator/cf-test-helpers/cf"
	"github.com/onsi/gomega/gbytes"
	"runtime"
)

var _ = Describe("hyperkit lifecycle", func() {
	var (
		startSession *gexec.Session
	)
	BeforeEach(func() {
		pluginPath = os.Getenv("CFDEV_PLUGIN_PATH")
		if pluginPath == "" {
			Fail("please provide CFDEV_PLUGIN_PATH (use ./generate-plugin.sh)")
		}
		os.Unsetenv("BOSH_ALL_PROXY")

		cfdevHome = os.Getenv("CFDEV_HOME")
		if cfdevHome == "" {
			cfdevHome = filepath.Join(os.Getenv("HOME"), ".cfdev")
		}
		hyperkitPidPath = filepath.Join(cfdevHome, "state", "linuxkit", "hyperkit.pid")

		fmt.Println("PLUGIN PATH: " + pluginPath)
		session := cf.Cf("install-plugin", pluginPath, "-f")
		Eventually(session).Should(gexec.Exit(0))
	})

	AfterEach(func() {
		//session := cf.Cf("uninstall-plugin", "cfdev")
		//Eventually(session).Should(gexec.Exit(0))
	})

	Context("starting the default cf dev file", func() {
		BeforeEach(func() {
			isoPath := os.Getenv("ISO_PATH")
			if isoPath != "" {
				startSession = cf.Cf("dev", "start", "-f", isoPath, "-m", "8192")
			} else {
				startSession = cf.Cf("dev", "start")
			}
		})

		AfterEach(func() {
			//if os.Getenv("CFDEV_FETCH_LOGS") == "true" {
			//	var logsSession *gexec.Session
			//
			//	if dir := os.Getenv("CFDEV_LOG_DIR"); dir != "" {
			//		logsSession = cf.Cf("dev", "logs", "--dir", dir)
			//	} else {
			//		logsSession = cf.Cf("dev", "logs")
			//	}
			//
			//	Eventually(logsSession).Should(gexec.Exit())
			//}
			//
			//hyperkitPid := PidFromFile(hyperkitPidPath)
			//
			//startSession.Terminate()
			//Eventually(startSession).Should(gexec.Exit())
			//
			//By("deploy finished - stopping...")
			//stopSession := cf.Cf("dev", "stop")
			//Eventually(stopSession).Should(gexec.Exit(0))
			//
			////ensure pid is not running
			//Eventually(IsLaunchdRunning("org.cloudfoundry.cfdev.linuxkit"), 5, 1).Should(BeFalse())
			//EventuallyProcessStops(hyperkitPid, 5)
			//Eventually(IsLaunchdRunning("org.cloudfoundry.cfdev.vpnkit"), 5, 1).Should(BeFalse())
			//
			//gexec.KillAndWait()
			//RemoveIPAliases(BoshDirectorIP, CFRouterIP)
		})

		FIt("runs the entire vm lifecycle", func() {
			Eventually(startSession, 20*time.Minute).Should(gbytes.Say("Starting VPNKit"))

			//daemonSpec := launchd.DaemonSpec{
			//	Label:"org.cloudfoundry.cfdev.vpnkit",
			//	CfDevHome: cfdevHome,
			//}
 
			Eventually(IsLaunchdRunning("org.cloudfoundry.cfdev.vpnkit"), 30, 1).Should(BeTrue())
			//Eventually(IsLaunchdRunning("org.cloudfoundry.cfdev.linuxkit"), 10, 1).Should(BeTrue())

			By("waiting for garden to listen")
			client := client.New(connection.New("tcp", "localhost:8888"))
			Eventually(client.Ping, 360).Should(Succeed())

			EventuallyWeCanTargetTheBOSHDirector()

			By("waiting for cfdev cli to exit when the deploy finished")
			Eventually(startSession, 3600).Should(gexec.Exit(0))

			By("waiting for cf router to listen")
			loginSession := cf.Cf("login", "-a", "https://api.v3.pcfdev.io", "--skip-ssl-validation", "-u", "admin", "-p", "admin", "-o", "cfdev-org", "-s", "cfdev-space")
			Eventually(loginSession).Should(gexec.Exit(0))
			//
			//By("pushing an app")
			//PushAnApp()
		})
	})

	Context("run with -f flag", func() {
		var assetDir string
		var startSession *gexec.Session

		BeforeEach(func() {
			var err error
			assetUrl := "https://s3.amazonaws.com/cfdev-test-assets/test-deps.dev"
			assetDir, err = ioutil.TempDir(os.TempDir(), "asset")
			Expect(err).ToNot(HaveOccurred())
			downloadTestAsset(assetDir, assetUrl)

			startSession = cf.Cf("dev", "start", "-f", filepath.Join(assetDir, "test-deps.dev"))
		})

		AfterEach(func() {
			hyperkitPid := PidFromFile(hyperkitPidPath)

			startSession.Terminate()
			Eventually(startSession).Should(gexec.Exit())

			session := cf.Cf("dev", "stop")
			Eventually(session).Should(gexec.Exit(0))

			//ensure pid is not running
			Eventually(IsLaunchdRunning("org.cloudfoundry.cfdev.linuxkit"), 5, 1).Should(BeFalse())
			EventuallyProcessStops(hyperkitPid, 5)
			Eventually(IsLaunchdRunning("org.cloudfoundry.cfdev.vpnkit"), 5, 1).Should(BeFalse())

			Expect(os.RemoveAll(assetDir)).To(Succeed())
		})

		It("Custom ISO", func() {
			Eventually(startSession, 20*time.Minute).Should(gbytes.Say("Starting VPNKit"))

			By("settingup VPNKit dependencies")
			Eventually(IsLaunchdRunning("org.cloudfoundry.cfdev.vpnkit"), 10, 1).Should(BeTrue())
			Eventually(IsLaunchdRunning("org.cloudfoundry.cfdev.linuxkit"), 10, 1).Should(BeTrue())

			By("waiting for garden to listen")
			EventuallyShouldListenAt("http://"+GardenIP+":8888", 360)

			client := client.New(connection.New("tcp", "localhost:8888"))
			Eventually(func() (string, error) {
				return GetFile(client, "deploy-bosh", "/var/vcap/cache/test-file-one.txt")
			}).Should(Equal("testfileone\n"))
		})
	})
})

func EventuallyWeCanTargetTheBOSHDirector() {
	By("waiting for bosh to listen")
	EventuallyShouldListenAt("https://"+BoshDirectorIP+":25555", 480)

	// Even though the test below is very similar this fails fast when `bosh env`
	// command is broken

	session := cf.Cf("dev", "bosh", "env")
	Eventually(session, 120, 1).Should(gexec.Exit(0))

	if runtime.GOOS != "windows" {
		// This test is more representative of how `bosh env` should be invoked
		w := gexec.NewPrefixedWriter("[bosh env] ", GinkgoWriter)
		boshEnv := func() *gexec.Session {
			boshCmd := exec.Command("/bin/sh",
				"-e",
				"-c", fmt.Sprintf(`eval "$(cf dev bosh env)" && bosh env`))

			session, err := gexec.Start(boshCmd, w, w)
			Expect(err).ToNot(HaveOccurred())
			Eventually(session).Should(gexec.Exit())
			return session
		}

		Eventually(boshEnv, time.Minute, 10*time.Second).Should(gexec.Exit(0))
	}
}

func RemoveIPAliases(aliases ...string) {
	if IsWindows() {
		return
	}

	for _, alias := range aliases {
		cmd := exec.Command("sudo", "-n", "ifconfig", "lo0", "inet", alias+"/32", "remove")
		writer := gexec.NewPrefixedWriter("[ifconfig] ", GinkgoWriter)
		session, err := gexec.Start(cmd, writer, writer)
		Expect(err).ToNot(HaveOccurred())
		Eventually(session).Should(gexec.Exit())
	}
}

func downloadTestAsset(targetDir string, resourceUrl string) error {
	out, err := os.Create(filepath.Join(targetDir, "test-deps.dev"))
	if err != nil {
		return err
	}
	defer out.Close()

	resp, err := http.Get(resourceUrl)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return err
	}

	return nil
}

func PushAnApp() {
	server, port := fakeTcpServer()
	defer server.Close()

	Eventually(cf.Cf("push", "cf-test-app", "--no-start", "-p", "./fixture", "-b", "ruby_buildpack"), 120).Should(gexec.Exit(0))
	Eventually(cf.Cf("set-env", "cf-test-app", "HOST_SERVER_PORT", strconv.Itoa(port)), 120).Should(gexec.Exit(0))
	Eventually(cf.Cf("create-service", "p-mysql", "10mb", "mydb"), 120).Should(gexec.Exit(0))
	Eventually(cf.Cf("bind-service", "cf-test-app", "mydb"), 120).Should(gexec.Exit(0))
	Eventually(cf.Cf("start", "cf-test-app"), 120).Should(gexec.Exit(0))

	Expect(httpGet("http://cf-test-app.v3.pcfdev.io")).To(Equal("Hello, world!"))
	Expect(httpGet("http://cf-test-app.v3.pcfdev.io/external")).To(ContainSubstring("Example Domain"))
	Expect(httpGet("http://cf-test-app.v3.pcfdev.io/host")).To(Equal("Text From Test Code"))
	Expect(httpGet("http://cf-test-app.v3.pcfdev.io/mysql")).To(ContainSubstring("innodb"))

	Eventually(cf.Cf("create-shared-domain", "tcp.v3.pcfdev.io", "--router-group", "default-tcp"), 10).Should(gexec.Exit(0))
	Eventually(cf.Cf("create-route", "cfdev-space", "tcp.v3.pcfdev.io", "--port", "1030"), 10).Should(gexec.Exit(0))
	Eventually(cf.Cf("map-route", "cf-test-app", "tcp.v3.pcfdev.io", "--port", "1030"), 10).Should(gexec.Exit(0))

	Eventually(func() (string, error) { return httpGet("http://tcp.v3.pcfdev.io:1030") }, 10).Should(Equal("Hello, world!"))
}

func fakeTcpServer() (net.Listener, int) {
	server, err := net.Listen("tcp", "0:0")
	Expect(err).NotTo(HaveOccurred())
	go func() {
		for {
			conn, err := server.Accept()
			if err != nil {
				continue
			}
			_, err = conn.Write([]byte("Text From Test Code"))
			Expect(err).NotTo(HaveOccurred())
			conn.Close()
		}
	}()
	return server, server.Addr().(*net.TCPAddr).Port
}

func httpGet(url string) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	b, err := ioutil.ReadAll(resp.Body)
	return string(b), err
}
