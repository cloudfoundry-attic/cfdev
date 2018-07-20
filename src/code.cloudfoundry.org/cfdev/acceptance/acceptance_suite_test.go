package acceptance

import (
	"io/ioutil"
	"path/filepath"
	"testing"

	"github.com/cloudfoundry-incubator/cf-test-helpers/cf"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"os"

	"time"

	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
	"runtime"
)

var (
	pluginPath                string
	tmpDir, cfHome, cfdevHome string
)

func TestCFDev(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "cf dev - acceptance suite")
}

var _ = BeforeSuite(func() {

	var err error

	pluginPath = os.Getenv("CFDEV_PLUGIN_PATH")
	if pluginPath == "" {
		if runtime.GOOS == "windows" {
			pluginPath, err = gexec.Build("code.cloudfoundry.org/cfdev", "-ldflags", `-X code.cloudfoundry.org/cfdev/config.cfdepsUrl=https://s3.amazonaws.com/cfdev-ci/cf-oss-deps/cf-deps-0.78.0.iso
     -X code.cloudfoundry.org/cfdev/config.cfdepsMd5=4a25c92fc4aa81d13e975049629b383d
     -X code.cloudfoundry.org/cfdev/config.cfdepsSize=4510406656
     -X code.cloudfoundry.org/cfdev/config.cfdevefiUrl=https://s3.amazonaws.com/cfdev-ci/cfdev-efi/cfdev-efi-windows-0.43.0.iso
     -X code.cloudfoundry.org/cfdev/config.cfdevefiMd5=9728fd7042772a9502093c4970f3a556
     -X code.cloudfoundry.org/cfdev/config.cfdevefiSize=342818816
     -X code.cloudfoundry.org/cfdev/config.vpnkitUrl=https://ci.appveyor.com/api/buildjobs/j8kg0kapwmcs3tq6/artifacts/vpnkit.exe
     -X code.cloudfoundry.org/cfdev/config.vpnkitMd5=dd97f6ce13069b4ed7291037e2c4ef8c
     -X code.cloudfoundry.org/cfdev/config.vpnkitSize=24514106
     -X code.cloudfoundry.org/cfdev/config.hyperkitUrl=https://s3.amazonaws.com/cfdev-ci/hyperkit/hyperkit-v0.20171204-24
     -X code.cloudfoundry.org/cfdev/config.hyperkitMd5=61da21b4e82e2bf2e752d043482aa966
     -X code.cloudfoundry.org/cfdev/config.hyperkitSize=3691536
     -X code.cloudfoundry.org/cfdev/config.linuxkitUrl=https://s3.amazonaws.com/cfdev-ci/linuxkit/linuxkit-darwin-amd64-0.0.0-build.10
     -X code.cloudfoundry.org/cfdev/config.linuxkitMd5=da8048c89e1cfa1f2a95ea27e83ae94c
     -X code.cloudfoundry.org/cfdev/config.linuxkitSize=44150800
     -X code.cloudfoundry.org/cfdev/config.winswUrl=https://s3.amazonaws.com/cfdev-ci/winsw/winsw.exe
     -X code.cloudfoundry.org/cfdev/config.winswMd5=1f41775fcf14aee2085c5fca5cd99d81
     -X code.cloudfoundry.org/cfdev/config.winswkitSize=360960
     -X code.cloudfoundry.org/cfdev/config.qcowtoolUrl=https://s3.amazonaws.com/cfdev-ci/qcow-tool/qcow-tool-v0.10.5
     -X code.cloudfoundry.org/cfdev/config.qcowtoolMd5=22f3a57096ae69027c13c4933ccdd96c
     -X code.cloudfoundry.org/cfdev/config.qcowtoolSize=4104388
     -X code.cloudfoundry.org/cfdev/config.uefiUrl=https://s3.amazonaws.com/cfdev-ci/uefi/UEFI-udk2014.sp1.fd
     -X code.cloudfoundry.org/cfdev/config.uefiMd5=2eff1c02d76fc3bde60f497ce1116b09
     -X code.cloudfoundry.org/cfdev/config.uefiSize=2097152
     -X code.cloudfoundry.org/cfdev/config.cfdevdUrl=https://s3.amazonaws.com/cfdev-ci/cfdevd/cfdevd-darwin-amd64-0.0.0-build.139
     -X code.cloudfoundry.org/cfdev/config.cfdevdMd5=27199b120ff884f331bbda7bc43a67b1
     -X code.cloudfoundry.org/cfdev/config.cfdevdSize=4721472
     -X code.cloudfoundry.org/cfdev/config.cliVersion=0.0.7-rc.49
     -X code.cloudfoundry.org/cfdev/config.analyticsKey=WFz4dVFXZUxN2Y6MzfUHJNWtlgXuOYV2
`)
		} else {
			pluginPath, err = gexec.Build("code.cloudfoundry.org/cfdev", "-ldflags", `-X code.cloudfoundry.org/cfdev/config.cfdepsUrl=https://s3.amazonaws.com/cfdev-ci/cf-oss-deps/cf-deps-0.78.0.iso
     -X code.cloudfoundry.org/cfdev/config.cfdepsMd5=4a25c92fc4aa81d13e975049629b383d
     -X code.cloudfoundry.org/cfdev/config.cfdepsSize=4510406656
     -X code.cloudfoundry.org/cfdev/config.cfdevefiUrl=https://s3.amazonaws.com/cfdev-ci/cfdev-efi/cfdev-efi-darwin-0.41.0.iso
     -X code.cloudfoundry.org/cfdev/config.cfdevefiMd5=3aee0faeda8c14ecb8536830ae76b612
     -X code.cloudfoundry.org/cfdev/config.cfdevefiSize=330307584
     -X code.cloudfoundry.org/cfdev/config.vpnkitUrl=https://s3.amazonaws.com/cfdev-ci/vpnkit/vpnkit-darwin-amd64-0.0.0-build.5
     -X code.cloudfoundry.org/cfdev/config.vpnkitMd5=045e7f7e4c97cbb4102ac836796f79e0
     -X code.cloudfoundry.org/cfdev/config.vpnkitSize=19655400
     -X code.cloudfoundry.org/cfdev/config.hyperkitUrl=https://s3.amazonaws.com/cfdev-ci/hyperkit/hyperkit-v0.20171204-24
     -X code.cloudfoundry.org/cfdev/config.hyperkitMd5=61da21b4e82e2bf2e752d043482aa966
     -X code.cloudfoundry.org/cfdev/config.hyperkitSize=3691536
     -X code.cloudfoundry.org/cfdev/config.linuxkitUrl=https://s3.amazonaws.com/cfdev-ci/linuxkit/linuxkit-darwin-amd64-0.0.0-build.10
     -X code.cloudfoundry.org/cfdev/config.linuxkitMd5=da8048c89e1cfa1f2a95ea27e83ae94c
     -X code.cloudfoundry.org/cfdev/config.linuxkitSize=44150800
     -X code.cloudfoundry.org/cfdev/config.qcowtoolUrl=https://s3.amazonaws.com/cfdev-ci/qcow-tool/qcow-tool-v0.10.5
     -X code.cloudfoundry.org/cfdev/config.qcowtoolMd5=22f3a57096ae69027c13c4933ccdd96c
     -X code.cloudfoundry.org/cfdev/config.qcowtoolSize=4104388
     -X code.cloudfoundry.org/cfdev/config.uefiUrl=https://s3.amazonaws.com/cfdev-ci/uefi/UEFI-udk2014.sp1.fd
     -X code.cloudfoundry.org/cfdev/config.uefiMd5=2eff1c02d76fc3bde60f497ce1116b09
     -X code.cloudfoundry.org/cfdev/config.uefiSize=2097152
     -X code.cloudfoundry.org/cfdev/config.cfdevdUrl=https://s3.amazonaws.com/cfdev-ci/cfdevd/cfdevd-darwin-amd64-0.0.0-build.139
     -X code.cloudfoundry.org/cfdev/config.cfdevdMd5=27199b120ff884f331bbda7bc43a67b1
     -X code.cloudfoundry.org/cfdev/config.cfdevdSize=4721472
     -X code.cloudfoundry.org/cfdev/config.cliVersion=0.0.7-rc.49
     -X code.cloudfoundry.org/cfdev/config.analyticsKey=WFz4dVFXZUxN2Y6MzfUHJNWtlgXuOYV2
`)
		}

		os.Setenv("CFDEV_PLUGIN_PATH", pluginPath)
	}

	Expect(err).ShouldNot(HaveOccurred())

	pluginPath = os.Getenv("CFDEV_PLUGIN_PATH")
	if pluginPath == "" {
		var err error
		pluginPath, err = gexec.Build("code.cloudfoundry.org/cfdev")
		Expect(err).ShouldNot(HaveOccurred())
	}

	SetDefaultEventuallyTimeout(5 * time.Second)
})

var _ = AfterSuite(func() {
	gexec.CleanupBuildArtifacts()
	Expect(os.RemoveAll(tmpDir)).To(Succeed())
})

var _ = BeforeEach(func() {
	var err error
	tmpDir, err = ioutil.TempDir("", "cfdev-acceptance-")
	Expect(err).ToNot(HaveOccurred())
	cfHome = filepath.Join(tmpDir, "cf_home")
	Expect(os.Mkdir(cfHome, 0755)).To(Succeed())
	os.Setenv("CF_HOME", cfHome)

	cfdevHome = filepath.Join(tmpDir, "cfdev_home")
	Expect(os.Mkdir(cfdevHome, 0755)).To(Succeed())
	os.Setenv("CFDEV_HOME", cfdevHome)

	session := cf.Cf("install-plugin", pluginPath, "-f")
	Eventually(session).Should(gexec.Exit(0))
	session = cf.Cf("plugins")
	Eventually(session).Should(gbytes.Say("cfdev"))
	Eventually(session).Should(gexec.Exit(0))
})

var _ = AfterEach(func() {
	session := cf.Cf("dev", "stop")
	Eventually(session, 30, 1).Should(gexec.Exit())
	os.Unsetenv("CF_HOME")
	os.Unsetenv("CFDEV_HOME")
})
