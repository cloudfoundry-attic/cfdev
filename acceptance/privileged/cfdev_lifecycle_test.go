package privileged_test

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/kinesis"
	"github.com/denisbrodbeck/machineid"
	"github.com/harlow/kinesis-consumer"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
	"io"
	"net"
	"os/exec"
	"runtime"
	"strconv"
	"time"

	"io/ioutil"
	"net/http"

	. "code.cloudfoundry.org/cfdev/acceptance"
	"github.com/cloudfoundry-incubator/cf-test-helpers/cf"
	"github.com/onsi/gomega/gbytes"
)

var _ = Describe("cfdev lifecycle", func() {

	var (
		analyticsChan = make(chan string, 50)
	)

	AfterEach(func() {
		if !cfg.CleanUp {
			fmt.Fprintln(GinkgoWriter, "'CLEANUP=false' environment variable detected, skipping tear-down...")
			return
		}

		greenColor := "\x1b[32;1m"
		endColor := "\x1b[0m"
		fmt.Fprintf(GinkgoWriter, "%s\n[STEP] Cleaning up...%s", greenColor, endColor)

		stopSession := cf.Cf("dev", "stop")
		Eventually(stopSession).Should(gexec.Exit(0))

		if runtime.GOOS == "windows" {
			Expect(doesVMExist()).To(BeFalse())
		}

		Eventually(IsServiceRunning("org.cloudfoundry.cfdev.linuxkit")).Should(BeFalse())

		if runtime.GOOS != "linux" {
			Eventually(IsServiceRunning("org.cloudfoundry.cfdev.vpnkit")).Should(BeFalse())
		}
	})

	It("runs the entire vm lifecycle", func() {
		startUp()

		go streamKinesis(analyticsChan)

		By("deploy mysql service")
		serviceSession := cf.Cf("dev", "deploy-service", "mysql")
		Eventually(serviceSession.Exited, 20*time.Minute).Should(BeClosed())

		By("waiting for cf router to listen")
		loginSession := cf.Cf("login", "-a", "https://api.dev.cfdev.sh", "--skip-ssl-validation", "-u", "admin", "-p", "admin", "-o", "cfdev-org", "-s", "cfdev-space")
		Eventually(loginSession).Should(gexec.Exit(0))

		By("toggling off telemetry")
		telemetrySession := cf.Cf("dev", "telemetry", "--off")
		Eventually(telemetrySession).Should(gexec.Exit(0))
		Eventually(IsServiceRunning("org.cloudfoundry.cfdev.cfanalyticsd")).Should(BeFalse())

		By("toggling telemetry on")
		telemetrySession = cf.Cf("dev", "telemetry", "--on")
		Eventually(telemetrySession).Should(gexec.Exit(0))
		Eventually(IsServiceRunning("org.cloudfoundry.cfdev.cfanalyticsd")).Should(BeTrue())

		// wait for analytics to perculate before
		// doing the rest of routine
		time.Sleep(20 * time.Second)

		By("pushing an app")
		PushAnApp()

		Expect(hasAnalyticsFor(analyticsChan, "app created", 3*time.Minute)).To(BeTrue())

		telemetrySession = cf.Cf("dev", "telemetry", "--off")
		Eventually(telemetrySession).Should(gexec.Exit(0))

		Expect(hasAnalyticsFor(analyticsChan, "telemetry off", 3*time.Minute)).To(BeTrue())

		By("rerunning cf dev start")
		startSession := cf.Cf("dev", "start")
		Eventually(startSession).Should(gbytes.Say("CF Dev is already running..."))

		By("checking for cf versions")
		var versionSession *gexec.Session

		if cfg.TarballPath != "" {
			versionSession = cf.Cf("dev", "version", "-f", cfg.TarballPath)
		} else {
			versionSession = cf.Cf("dev", "version")
		}

		Eventually(versionSession).Should(gexec.Exit(0))
		Expect(string(versionSession.Out.Contents())).To(ContainSubstring("CLI:"))
		Expect(string(versionSession.Out.Contents())).To(ContainSubstring("mysql:"))
	})
})

func startUp() {
	if !cfg.StartUp {
		fmt.Fprintln(GinkgoWriter, "'STARTUP=false' environment variable detected, skipping start up...")
		return
	}

	var startSession *gexec.Session

	// stop should succeed even when nothing is running
	stopSession := cf.Cf("dev", "stop")
	Eventually(stopSession).Should(gexec.Exit(0))

	if cfg.TarballPath != "" {
		startSession = cf.Cf("dev", "start", "-f", cfg.TarballPath)
	} else {
		startSession = cf.Cf("dev", "start")
	}

	By("waiting for bosh to deploy")
	Eventually(startSession, 2*time.Hour).Should(gbytes.Say("Deploying the BOSH Director"))

	EventuallyWeCanTargetTheBOSHDirector()

	By("waiting for cfdev cli to exit when the deploy finished")
	Eventually(startSession.Exited, 2*time.Hour).Should(BeClosed())
	Expect(startSession.ExitCode()).To(BeZero())
}

func hasAnalyticsFor(analyticsChan chan string, eventName string, timeout time.Duration) bool {
	if cfg.AwsAccessKeyID == "" || cfg.AwsSecretAccessKey == "" {
		return true
	}

	timeoutChan := time.After(timeout)
	By(fmt.Sprintf("Waiting for analytics `%s` to be received", eventName))

	for {
		select {
		case <-timeoutChan:
			return false
		case element := <-analyticsChan:
			if element == eventName {
				return true
			}
		}
	}
}

func EventuallyWeCanTargetTheBOSHDirector() {
	By("waiting for bosh to listen")

	Eventually(func() error {
		return HttpServerIsListeningAt("https://" + BoshDirectorIP + ":25555")
	}, 15*time.Minute, 30*time.Second).ShouldNot(HaveOccurred())

	w := gexec.NewPrefixedWriter("[bosh env] ", GinkgoWriter)
	Eventually(func() error {

		var boshCmd *exec.Cmd

		if runtime.GOOS == "windows" {
			boshCmd = exec.Command("powershell.exe",
				"-Command",
				`cf dev bosh env | Invoke-Expression; bosh env`)
		} else {
			boshCmd = exec.Command("/bin/sh",
				"-e",
				"-c", `eval "$(cf dev bosh env)" && bosh env`)
		}

		output, err := boshCmd.CombinedOutput()
		fmt.Fprintln(w, string(output))
		return err
	}, 10*time.Minute, 30*time.Second).Should(BeNil())
}

func PushAnApp() {
	port := "36167"
	srv := &http.Server{Addr: ":"+port}
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		io.WriteString(w, "Text From Test Code")
	})

	go srv.ListenAndServe()
	defer srv.Shutdown(context.TODO())

	Eventually(cf.Cf("push", "cf-test-app", "--no-start", "-p", "./fixture", "-b", "ruby_buildpack")).Should(gexec.Exit(0))
	Eventually(cf.Cf("set-env", "cf-test-app", "HOST_SERVER_PORT", port)).Should(gexec.Exit(0))
	Eventually(cf.Cf("create-service", cfg.MysqlService, cfg.MysqlServicePlan, "mydb")).Should(gexec.Exit(0))

	Eventually(func() string {
		sesh := cf.Cf("service", "mydb")
		<-sesh.Exited
		return string(sesh.Out.Contents())
	}, 30*time.Minute, 5*time.Second).Should(ContainSubstring("create succeeded"))

	Eventually(cf.Cf("bind-service", "cf-test-app", "mydb")).Should(gexec.Exit(0))

	Eventually(func() int {
		sesh := cf.Cf("start", "cf-test-app")
		<-sesh.Exited
		return sesh.ExitCode()
	}, 10*time.Minute, 10*time.Second).Should(BeZero())

	Expect(httpGet("http://cf-test-app.dev.cfdev.sh")).To(Equal("Hello, world!"))
	Expect(httpGet("http://cf-test-app.dev.cfdev.sh/external")).To(ContainSubstring("Example Domain"))
	Expect(httpGet("http://cf-test-app.dev.cfdev.sh/host")).To(Equal("Text From Test Code"))
	Expect(httpGet("http://cf-test-app.dev.cfdev.sh/mysql")).To(ContainSubstring("innodb"))

	Eventually(cf.Cf("create-shared-domain", "tcp.dev.cfdev.sh", "--router-group", "default-tcp")).Should(gexec.Exit(0))
	Eventually(cf.Cf("create-route", "cfdev-space", "tcp.dev.cfdev.sh", "--port", "1030")).Should(gexec.Exit(0))
	Eventually(cf.Cf("map-route", "cf-test-app", "tcp.dev.cfdev.sh", "--port", "1030")).Should(gexec.Exit(0))

	Eventually(func() (string, error) {
		return httpGet("http://tcp.dev.cfdev.sh:1030")
	}).Should(Equal("Hello, world!"))
}

func fakeTcpServer() (net.Listener, int) {
	server, err := net.Listen("tcp", "localhost:0")
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

func doesVMExist() bool {
	cmd := exec.Command("powershell.exe", "-Command", "(Get-VM -Name cfdev).name")
	output, err := cmd.Output()
	if err != nil {
		return false
	}

	return string(output) == "cfdev"
}

type StatMessage struct {
	UserId    string `json:"userId"`
	Event     string `json:"event"`
	Timestamp string `json:"timestamp"`
}

func streamKinesis(analyticsChan chan string) {
	if cfg.AwsAccessKeyID == "" || cfg.AwsSecretAccessKey == "" {
		fmt.Fprintln(GinkgoWriter, "AWS keys not detected. Skipping assertions for analytics...")
		return
	}

	userID, _ := machineid.ProtectedID("cfdev")
	stream := flag.String("cfdev-analytics-development", "cfdev-analytics-development", "cfdev-analytics-development")
	flag.Parse()

	myKinesisClient := kinesis.New(session.New(aws.NewConfig()), &aws.Config{
		Region:      aws.String("us-east-1"),
		Credentials: credentials.NewStaticCredentials(cfg.AwsAccessKeyID, cfg.AwsSecretAccessKey, ""),
	})
	newKclient, err := consumer.NewKinesisClient(consumer.WithKinesis(myKinesisClient))
	c, err := consumer.New(
		*stream,
		consumer.WithClient(newKclient),
	)
	if err != nil {
		fmt.Printf("consumer error: %v \n", err)
	}
	ctx, _ := context.WithCancel(context.Background())
	err = c.Scan(ctx, func(r *consumer.Record) consumer.ScanError {
		var analyticsEvent StatMessage
		json.Unmarshal(r.Data, &analyticsEvent)
		eventTime, err := time.Parse(time.RFC3339, analyticsEvent.Timestamp)
		tenMinutesAgo := time.Now().UTC().Add(-10 * time.Minute)
		if eventTime.After(tenMinutesAgo) {
			if analyticsEvent.UserId == userID {
				analyticsChan <- analyticsEvent.Event
			}
		}
		err = errors.New("some error happened")

		return consumer.ScanError{
			Error:          err,
			StopScan:       false,
			SkipCheckpoint: false,
		}
	})
}
