package main

import (
	"code.cloudfoundry.org/cfdev/analyticsd/daemon"
	"context"
	"crypto/tls"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"
	"gopkg.in/segmentio/analytics-go.v3"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
	"github.com/denisbrodbeck/machineid"
)

func main() {
	var (
		analyticsKey string
		userID string
		clientSecret string
		tokenUrl string
	)

	cfg := &clientcredentials.Config{
		ClientID:     "cfdev_analytics",
		ClientSecret: clientSecret,
		TokenURL:     tokenUrl,
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
		os.Stdout,
		cfg.Client(ctx),
		analytics.New(analyticsKey),
		10*time.Minute,
		time.Time{},
	)

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigs
		analyticsDaemon.Stop()
	}()

	analyticsDaemon.Start()
}