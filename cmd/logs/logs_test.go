package logs_test

import (
	"os"
	"path/filepath"

	"code.cloudfoundry.org/cfdev/cmd/logs"
	"code.cloudfoundry.org/cfdev/cmd/logs/mocks"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Logs", func() {
	var (
		mockController  *gomock.Controller
		mockProvisioner *mocks.MockProvisioner
		mockUI          *mocks.MockUI
	)

	Describe("Logs", func() {
		var cmd *logs.Logs

		BeforeEach(func() {
			mockController = gomock.NewController(GinkgoT())
			mockProvisioner = mocks.NewMockProvisioner(mockController)
			mockUI = mocks.NewMockUI(mockController)

			cmd = &logs.Logs{
				Provisioner: mockProvisioner,
				UI:          mockUI,
			}
		})

		AfterEach(func() {
			mockController.Finish()
		})

		It("fetches logs", func() {
			mockProvisioner.EXPECT().FetchLogs("some-dir")
			wd, err := os.Getwd()
			Expect(err).NotTo(HaveOccurred())
			logPath := filepath.Join(wd, "some-dir", "cfdev-logs.tgz")
			mockUI.EXPECT().Say("Logs downloaded to " + logPath)

			Expect(cmd.Logs(logs.Args{
				DestDir: "some-dir",
			})).To(Succeed())
		})
	})
})
