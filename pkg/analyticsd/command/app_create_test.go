package command_test

import (
	"code.cloudfoundry.org/cfdev/pkg/analyticsd/command"
	"code.cloudfoundry.org/cfdev/pkg/analyticsd/segment"
	"code.cloudfoundry.org/cfdev/pkg/analyticsd/segment/mocks"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"gopkg.in/segmentio/analytics-go.v3"
	"io/ioutil"
	"log"
	"time"
)

var _ = Describe("AppCreate", func() {
	var (
		cmd            *command.AppCreate
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

		cmd = &command.AppCreate{
			Logger:          log.New(ioutil.Discard, "", log.LstdFlags),
			AnalyticsClient: segmentClient,
		}
	})

	AfterEach(func() {
		mockController.Finish()
	})

	Context("when the buildpack is whitelisted", func() {
		It("sends the buildpack information to segment.io", func() {
			mockAnalytics.EXPECT().Enqueue(gomock.Any()).Do(func(event analytics.Track) {
				Expect(event.UserId).To(Equal("some-user-uuid"))
				Expect(event.Event).To(Equal("app created"))
				Expect(event.Timestamp).To(Equal(time.Date(2018, 8, 8, 8, 8, 8, 0, time.UTC)))
				Expect(event.Properties).To(HaveKeyWithValue("buildpack", "go"))
			})

			body := []byte(`
			{
				"request": {
					"buildpack": "go_buildpack"
				}
			}`)

			cmd.HandleResponse(body)
		})
	})

	Context("when the buildpack is not whitelisted", func() {
		It("sends the buildpack information to segment.io", func() {
			mockAnalytics.EXPECT().Enqueue(gomock.Any()).Do(func(event analytics.Track) {
				Expect(event.UserId).To(Equal("some-user-uuid"))
				Expect(event.Event).To(Equal("app created"))
				Expect(event.Timestamp).To(Equal(time.Date(2018, 8, 8, 8, 8, 8, 0, time.UTC)))
				Expect(event.Properties).To(HaveKeyWithValue("buildpack", "custom"))
			})

			body := []byte(`
			{
				"request": {
					"buildpack": "some-unexpected-buildpack"
				}
			}`)

			cmd.HandleResponse(body)
		})
	})
})
