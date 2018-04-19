package cmd

import (
	"fmt"
	"os"
	"strings"

	"io/ioutil"
	"path/filepath"

	"code.cloudfoundry.org/cfdev/config"
	"code.cloudfoundry.org/cfdev/env"
	"code.cloudfoundry.org/cfdev/resource"
	"github.com/spf13/cobra"
)

func NewDownload(Exit chan struct{}, UI UI, Config config.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use: "download",
		RunE: func(cmd *cobra.Command, args []string) error {
			go func() {
				<-Exit
				os.Exit(128)
			}()

			if err := env.Setup(Config); err != nil {
				return nil
			}

			UI.Say("Downloading Resources...")
			return download(Config.Dependencies, Config.CacheDir)
		},
	}
	return cmd
}

func download(dependencies resource.Catalog, cacheDir string) error {
	logCatalog(dependencies, cacheDir)
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

func logCatalog(dependencies resource.Catalog, cacheDir string) {
	ioutil.WriteFile(filepath.Join(cacheDir, "catalog.txt"), []byte(fmt.Sprintf("%+v", dependencies.Items)), 0644)
}
