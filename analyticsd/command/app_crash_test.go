package command_test

import (
	"code.cloudfoundry.org/cfdev/analyticsd/command"
	"code.cloudfoundry.org/cfdev/analyticsd/segment"
	"code.cloudfoundry.org/cfdev/analyticsd/segment/mocks"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"gopkg.in/segmentio/analytics-go.v3"
	"io/ioutil"
	"log"
	"time"
)

var _ = Describe("AppCrash", func() {
	var (
		cmd            *command.AppCrash
		mockController *gomock.Controller
		mockAnalytics  *mocks.MockClient
	)

	BeforeEach(func() {
		mockController = gomock.NewController(GinkgoT())
		mockAnalytics = mocks.NewMockClient(mockController)
		segmentClient := segment.New(
			mockAnalytics,
			"some-user-uuid",
			"some-version",
			"some-os-version",
			time.Date(2018, 8, 8, 8, 8, 8, 0, time.UTC),
		)

		cmd = &command.AppCrash{
			Logger:          log.New(ioutil.Discard, "", log.LstdFlags),
			AnalyticsClient: segmentClient,
		}
	})

	AfterEach(func() {
		mockController.Finish()
	})

	Context("when an app crash event occurs", func() {
		It("sends the the crash event to segment.io", func() {
			mockAnalytics.EXPECT().Enqueue(gomock.Any()).Do(func(event analytics.Track) {
				Expect(event.UserId).To(Equal("some-user-uuid"))
				Expect(event.Event).To(Equal("app push failed"))
				Expect(event.Timestamp).To(Equal(time.Date(2018, 8, 8, 8, 8, 8, 0, time.UTC)))
			})

			body := []byte(`
			{
                  "instance": "d73d6816-3101-4efb-4a9b-c4c1",
                  "index": 0,
                  "cell_id": "f5002113-527c-4bf9-b5e4-867f68a4ecfc",
                  "exit_description": "APP/PROC/WEB: Exited with status 1",
                  "reason": "CRASHED"
            }`)

			cmd.HandleResponse(body)
		})
	})
})
