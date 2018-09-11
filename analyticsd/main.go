package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"code.cloudfoundry.org/cfdev/analyticsd/daemon"
	"github.com/denisbrodbeck/machineid"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"
	"gopkg.in/segmentio/analytics-go.v3"
)

var (
	analyticsKey    string
	version         string
	pollingInterval = 10*time.Minute
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

	if len(os.Args) > 1 && os.Args[1] == "debug" {
		pollingInterval = 10*time.Second
		fmt.Printf("[DEBUG] analyticsKey: %q\n", analyticsKey)
		fmt.Printf("[DEBUG] userID: %q\n", userID)
		fmt.Printf("[DEBUG] pollingInterval: %v\n", pollingInterval)
		fmt.Printf("[DEBUG] version %q\n", version)
	}

	analyticsDaemon := daemon.New(
		"https://api.dev.cfdev.sh",
		userID,
		version,
		os.Stdout,
		cfg.Client(ctx),
		analytics.New(analyticsKey),
		pollingInterval,
	)

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigs
		analyticsDaemon.Stop()
	}()

	analyticsDaemon.Start()
}
