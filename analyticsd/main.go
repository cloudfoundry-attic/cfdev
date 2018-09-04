package main

import (
	"context"
	"crypto/tls"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"code.cloudfoundry.org/cfdev/analyticsd/daemon"
	"github.com/denisbrodbeck/machineid"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"
	analytics "gopkg.in/segmentio/analytics-go.v3"
)

var (
	analyticsKey string
	version      string
)

func main() {
	cfg := &clientcredentials.Config{
		ClientID:     "analytics",
		ClientSecret: "analytics",
		TokenURL:     "https://uaa.dev.cfdev.sh/oauth/token",
	}

	httpClient := &http.Client{
		Timeout: 10 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
	}

	ctx := context.Background()
	ctx = context.WithValue(ctx, oauth2.HTTPClient, httpClient)

	userID, err := machineid.ProtectedID("cfdev")
	if err != nil {
		userID = "UNKNOWN_ID"
	}

	analyticsDaemon := daemon.New(
		"https://api.dev.cfdev.sh",
		userID,
		version,
		os.Stdout,
		cfg.Client(ctx),
		analytics.New(analyticsKey),
		10*time.Minute,
	)

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigs
		analyticsDaemon.Stop()
	}()

	analyticsDaemon.Start()
}
