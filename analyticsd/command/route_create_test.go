package command_test

import (
	"code.cloudfoundry.org/cfdev/analyticsd/command"
	commandMocks "code.cloudfoundry.org/cfdev/analyticsd/command/mocks"
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

var _ = Describe("RouteCreate", func() {
	var (
		cmd            *command.RouteCreate
		mockController *gomock.Controller
		mockAnalytics  *mocks.MockClient
		mockCCClient   *commandMocks.MockCloudControllerClient
	)

	BeforeEach(func() {
		mockController = gomock.NewController(GinkgoT())
		mockAnalytics = mocks.NewMockClient(mockController)
		mockCCClient = commandMocks.NewMockCloudControllerClient(mockController)
		segmentClient := segment.New(
			mockAnalytics,
			"some-user-uuid",
			"some-version",
			"some-os-version",
			time.Date(2018, 8, 8, 8, 8, 8, 0, time.UTC),
		)

		cmd = &command.RouteCreate{
			Logger:          log.New(ioutil.Discard, "", log.LstdFlags),
			CCClient:        mockCCClient,
			AnalyticsClient: segmentClient,
		}
	})

	AfterEach(func() {
		mockController.Finish()
	})

	Context("when route is created", func() {
		It("sends the route information to segment.io", func() {
			mockAnalytics.EXPECT().Enqueue(gomock.Any()).Do(func(event analytics.Track) {
				Expect(event.UserId).To(Equal("some-user-uuid"))
				Expect(event.Event).To(Equal("created route"))
				Expect(event.Timestamp).To(Equal(time.Date(2018, 8, 8, 8, 8, 8, 0, time.UTC)))
			})


			body := []byte("")

			Expect(cmd.HandleResponse(body)).NotTo(HaveOccurred())
		})
	})
})
