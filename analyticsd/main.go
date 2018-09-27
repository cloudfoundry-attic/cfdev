package main

import (
	"code.cloudfoundry.org/cfdev/analyticsd/runner"
	"context"
	"crypto/tls"
	"fmt"
	"gopkg.in/segmentio/analytics-go.v3"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"code.cloudfoundry.org/cfdev/analyticsd/daemon"
	analyticsdos "code.cloudfoundry.org/cfdev/analyticsd/os"
	"github.com/denisbrodbeck/machineid"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"
)

var (
	analyticsKey    string
	version         string
	pollingInterval = 10 * time.Minute
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

	analyticsdos :=  &analyticsdos.OS{Runner: &runner.Runner{}}
	osVersion, err := analyticsdos.Version()
	if err != nil {
		osVersion = "unknown-os-version"
	}

	if len(os.Args) > 1 && os.Args[1] == "debug" {
		pollingInterval = 10 * time.Second
	}

	analyticsDaemon := daemon.New(
		"https://api.dev.cfdev.sh",
		userID,
		version,
		osVersion,
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

	fmt.Printf("[ANALYTICSD] apiKeyLoaded: %t, pollingInterval: %v, version: %q, time: %v, userID: %q\n",
		analyticsKey != "", pollingInterval, version, time.Now(), userID)
	analyticsDaemon.Start()
}
