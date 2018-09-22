package command_test

import (
	"code.cloudfoundry.org/cfdev/analyticsd/command/mocks"
	"encoding/json"
	"github.com/golang/mock/gomock"
	"net/url"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestCommand(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Command Suite")
}

func MatchFetch(mockCCClient *mocks.MockCloudControllerClient, expectedPath, result string) {
	mockCCClient.EXPECT().Fetch(gomock.Any(), gomock.Any(), gomock.Any()).Do(func(path string, params url.Values, dest interface{}) error {
		Expect(path).To(Equal(expectedPath))
		return json.Unmarshal([]byte(result), dest)
	})
}