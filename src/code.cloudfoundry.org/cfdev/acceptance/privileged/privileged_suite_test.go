package privileged_test

import (
	"os/exec"
	"testing"
	"time"

	. "code.cloudfoundry.org/cfdev/acceptance"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
	"runtime"
)

var pluginPath string
var cfdevHome string
var hyperkitPidPath string

func TestPrivileged(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "cf dev - acceptance - privileged suite")
}

var _ = BeforeSuite(func() {
	SetDefaultEventuallyTimeout(10 * time.Second)

	Expect(HasSudoPrivilege()).To(BeTrue(), "Please run 'sudo echo hi' first")
	RemoveIPAliases(BoshDirectorIP, CFRouterIP)

	//TODO make sure you want this
	var err error
	pluginPath, err = gexec.Build("code.cloudfoundry.org/cfdev","-ldflags", `-X code.cloudfoundry.org/cfdev/config.cfdepsUrl=https://s3.amazonaws.com/cfdev-ci/cf-oss-deps/cf-deps-0.78.0.iso
     -X code.cloudfoundry.org/cfdev/config.cfdepsMd5=4a25c92fc4aa81d13e975049629b383d
     -X code.cloudfoundry.org/cfdev/config.cfdepsSize=4510406656
     -X code.cloudfoundry.org/cfdev/config.cfdevefiUrl=https://s3.amazonaws.com/cfdev-ci/cfdev-efi/cfdev-efi-darwin-0.41.0.iso
     -X code.cloudfoundry.org/cfdev/config.cfdevefiMd5=3aee0faeda8c14ecb8536830ae76b612
     -X code.cloudfoundry.org/cfdev/config.cfdevefiSize=330307584
     -X code.cloudfoundry.org/cfdev/config.vpnkitUrl=C:\\Users\\WX014\\Desktop\\cfdev\\src\\code.cloudfoundry.org\\cfdev\\vpnkit.exe
     -X code.cloudfoundry.org/cfdev/config.vpnkitMd5=f1ecd9b3b2d91983cb41b55302de09bb
     -X code.cloudfoundry.org/cfdev/config.vpnkitSize=24349961
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
	Expect(err).ShouldNot(HaveOccurred())
})

var _ = AfterSuite(func() {
	gexec.CleanupBuildArtifacts()
})

func HasSudoPrivilege() bool {
	if IsWindows() {
		return true
	}

	cmd := exec.Command("sh", "-c", "sudo -n true")
	session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
	Expect(err).NotTo(HaveOccurred())
	Eventually(session).Should(gexec.Exit())

	if session.ExitCode() == 0 {
		return true
	}
	return false
}

func IsWindows() bool {
	return runtime.GOOS == "windows"
}
