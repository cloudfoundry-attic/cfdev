package bosh_test

import (
	"errors"

	"code.cloudfoundry.org/cfdev/bosh"
	"code.cloudfoundry.org/cfdev/bosh/mocks"
	boshdir "github.com/cloudfoundry/bosh-cli/director"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

//go:generate mockgen -package mocks -destination mocks/director.go github.com/cloudfoundry/bosh-cli/director Director
//go:generate mockgen -package mocks -destination mocks/deployment.go github.com/cloudfoundry/bosh-cli/director Deployment

var _ = Describe("Bosh", func() {
	var (
		subject        *bosh.Bosh
		mockController *gomock.Controller
		mockDir        *mocks.MockDirector
		mockDep        *mocks.MockDeployment
	)
	BeforeEach(func() {
		bosh.VMProgressInterval = 0
		mockController = gomock.NewController(GinkgoT())
		mockDir = mocks.NewMockDirector(mockController)
		mockDep = mocks.NewMockDeployment(mockController)
		subject = bosh.NewWithDirector(mockDir)
	})
	AfterEach(func() {
		mockController.Finish()
	})

	Describe("VMProgress", func() {
		It("swallows finding cf errors, returns releases until vms, and then vms", func() {
			mockDir.EXPECT().FindDeployment("cf").Return(nil, errors.New("not found"))
			mockDir.EXPECT().FindDeployment("cf").Return(nil, errors.New("not found"))
			mockDir.EXPECT().FindDeployment("cf").Return(mockDep, nil)
			mockDir.EXPECT().Releases().Return([]boshdir.Release{}, nil)
			mockDir.EXPECT().Releases().AnyTimes().Return([]boshdir.Release{nil, nil}, nil)
			vmInfos := []boshdir.VMInfo{}
			mockDepVMInfos := mockDep.EXPECT().VMInfos().AnyTimes()
			mockDepVMInfos.Do(func() {
				mockDepVMInfos.Return(vmInfos, nil)
			})

			ch := subject.VMProgress()
			var p bosh.VMProgress

			Eventually(ch).Should(Receive(&p))
			Expect(p.Releases).To(Equal(0))

			Eventually(ch).Should(Receive(&p))
			Expect(p.Releases).To(Equal(2))

			vmInfos = []boshdir.VMInfo{
				boshdir.VMInfo{},
			}

			Eventually(func() []int {
				p := <-ch
				return []int{p.Releases, p.Total, p.Done}
			}).Should(Equal([]int{0, 1, 0}))

			vmInfos = []boshdir.VMInfo{
				boshdir.VMInfo{ProcessState: "queued", Processes: []boshdir.VMInfoProcess{}},
				boshdir.VMInfo{ProcessState: "running", Processes: []boshdir.VMInfoProcess{
					boshdir.VMInfoProcess{},
				}},
				boshdir.VMInfo{ProcessState: "running", Processes: []boshdir.VMInfoProcess{}},
			}

			Eventually(func() []int {
				p := <-ch
				return []int{p.Releases, p.Total, p.Done}
			}).Should(Equal([]int{0, 3, 1}))
		})
	})
})
