package cmd

import (
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"io/ioutil"
	"path/filepath"

	"net/http"

	"code.cloudfoundry.org/cfdev/config"
	"code.cloudfoundry.org/cfdev/env"
	"code.cloudfoundry.org/cfdev/errors"
	"code.cloudfoundry.org/cfdev/resource"
	"code.cloudfoundry.org/cfdev/resource/progress"
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
				return errors.SafeWrap(err, "setup for download")
			}

			UI.Say("Downloading Resources...")
			return download(Config.Dependencies, Config.CacheDir, UI.Writer())
		},
	}
	return cmd
}

func download(dependencies resource.Catalog, cacheDir string, writer io.Writer) error {
	logCatalog(dependencies, cacheDir)
	skipVerify := strings.ToLower(os.Getenv("CFDEV_SKIP_ASSET_CHECK"))

	cache := resource.Cache{
		Dir:                   cacheDir,
		HttpDo:                http.DefaultClient.Do,
		SkipAssetVerification: skipVerify == "true",
		Progress:              progress.New(writer),
		RetryWait:             time.Second,
		Writer:                writer,
	}

	if err := cache.Sync(&dependencies); err != nil {
		return errors.SafeWrap(err, "Unable to sync assets")
	}
	return nil
}

func logCatalog(dependencies resource.Catalog, cacheDir string) {
	ioutil.WriteFile(filepath.Join(cacheDir, "catalog.txt"), []byte(fmt.Sprintf("%+v", dependencies.Items)), 0644)
}
