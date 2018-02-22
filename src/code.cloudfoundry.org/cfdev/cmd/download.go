package cmd

import (
	"fmt"
	"code.cloudfoundry.org/cfdev/resource"
	"strings"
	"os"
	"code.cloudfoundry.org/cfdev/config"
	"code.cloudfoundry.org/cfdev/env"
)

type Download struct {
	Exit chan struct{}
	UI UI
	Config config.Config
}

func (d *Download) Run(args []string) error {
	go func() {
		<-d.Exit
		os.Exit(128)
	}()

	if err := env.Setup(d.Config); err != nil {
		return nil
	}

	d.UI.Say("Downloading Resources...")
	return download(d.Config.Dependencies, d.Config.CacheDir)
}

func download(dependencies resource.Catalog, cacheDir string) error {
	downloader := resource.Downloader{}
	skipVerify := strings.ToLower(os.Getenv("CFDEV_SKIP_ASSET_CHECK"))

	cache := resource.Cache{
		Dir:                   cacheDir,
		DownloadFunc:          downloader.Start,
		SkipAssetVerification: skipVerify == "true",
	}

	if err := cache.Sync(&dependencies); err != nil {
		return fmt.Errorf("Unable to sync assets: %v\n", err)
	}
	return nil
}