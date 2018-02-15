package cmd

import (
	"fmt"
	"code.cloudfoundry.org/cfdev/resource"
	"strings"
	"os"
)

type Download struct {
	Exit chan struct{}
	UI UI
}

func (d *Download) Run(args []string) error {
	go func() {
		<-d.Exit
		os.Exit(128)
	}()

	_, _, cacheDir, err := setupHomeDir()
	if err != nil {
		return nil
	}

	d.UI.Say("Downloading Resources...")
	downloader := resource.Downloader{}
	skipVerify := strings.ToLower(os.Getenv("CFDEV_SKIP_ASSET_CHECK"))

	cache := resource.Cache{
		Dir:                   cacheDir,
		DownloadFunc:          downloader.Start,
		SkipAssetVerification: skipVerify == "true",
	}

	catalog, err := catalog(d.UI)
	if err != nil {
		return err
	}

	if err := cache.Sync(catalog); err != nil {
		return fmt.Errorf("Unable to sync assets: %v\n", err)
	}
	return nil
}
