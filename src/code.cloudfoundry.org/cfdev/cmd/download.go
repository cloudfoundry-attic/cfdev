package cmd

import (
	"fmt"
	"code.cloudfoundry.org/cfdev/resource"
	"strings"
	"os"
)

type Download struct{
	Exit chan struct{}
}

func(d *Download) Run(args []string) {
	go func() {
		<-d.Exit
		os.Exit(128)
	}()

	_, _, cacheDir := setupHomeDir()
	fmt.Println("Downloading Resources...")
	downloader := resource.Downloader{}
	skipVerify := strings.ToLower(os.Getenv("CFDEV_SKIP_ASSET_CHECK"))

	cache := resource.Cache{
		Dir:                   cacheDir,
		DownloadFunc:          downloader.Start,
		SkipAssetVerification: skipVerify == "true",
	}

	if err := cache.Sync(catalog()); err != nil {
		fmt.Fprintf(os.Stderr, "Unable to sync assets: %v\n", err)
		os.Exit(1)
	}
}