package privileged_test

import (
	. "code.cloudfoundry.org/cfdev/acceptance"
	"encoding/json"
	"fmt"
	"github.com/cloudfoundry-incubator/cf-test-helpers/cf"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"io/ioutil"
	"os"
	"testing"
	"time"
)

type config struct {
	AwsAccessKeyID     string `json:"aws_access_key_id"`
	AwsSecretAccessKey string `json:"aws_secret_access_key"`
	StartUp            bool   `json:"start_up"`
	CleanUp            bool   `json:"clean_up"`
	TarballPath        string `json:"tarball_path"`
	PluginPath         string `json:"plugin_path"`

	MysqlService     string `json:"service"`
	MysqlServicePlan string `json:"service_plan"`
}

var cfg = config{
	StartUp:          true,
	CleanUp:          true,
	MysqlService:     "p-mysql",
	MysqlServicePlan: "10mb",
}

func TestPrivileged(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "cf dev - acceptance - privileged suite")
}

var _ = BeforeSuite(func() {
	SetDefaultEventuallyTimeout(5 * time.Minute)

	Expect(HasSudoPrivilege()).To(BeTrue(), "Please run 'sudo echo hi' first")

	if path := os.Getenv("CFDEV_ACCEPTANCE_CONFIG"); path != "" {
		contents, err := ioutil.ReadFile(path)
		if err != nil {
			Fail(fmt.Sprintf("failed to read acceptance config file at: %s", path))
		}

		err = json.Unmarshal([]byte(contents), &cfg)
		if err != nil {
			Fail(fmt.Sprintf("Unable to parse 'CFDEV_ACCEPTANCE_CONFIG'. Please make sure that the env var is valid json: %s", contents))
		}
	}

	if cfg.PluginPath != "" {
		session := cf.Cf("install-plugin", "-f", cfg.PluginPath)
		<-session.Exited
	} else {
		fmt.Fprintln(GinkgoWriter, "WARNING plugin_path omitted as part of 'CFDEV_ACCEPTANCE_CONFIG'. Skipping plugin installation...")
	}

	os.Setenv("CFDEV_MODE", "debug")
	os.Setenv("CF_COLOR", "false")
	os.Unsetenv("BOSH_ALL_PROXY")
})

var _ = AfterSuite(func() {
	os.Unsetenv("CF_COLOR")
	os.Unsetenv("CFDEV_MODE")

	if cfg.PluginPath != "" {
		cf.Cf("uninstall-plugin", "cfdev")
	}
})
