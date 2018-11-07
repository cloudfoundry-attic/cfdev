package download

import (
	"io"
	"os"
	"strings"
	"time"

	"net/http"

	"code.cloudfoundry.org/cfdev/config"
	"code.cloudfoundry.org/cfdev/errors"
	"code.cloudfoundry.org/cfdev/resource"
	"code.cloudfoundry.org/cfdev/resource/progress"
	"github.com/spf13/cobra"
)

type UI interface {
	Say(message string, args ...interface{})
	Writer() io.Writer
}

//go:generate mockgen -package mocks -destination mocks/env.go code.cloudfoundry.org/cfdev/cmd/start Env
type Env interface {
	CreateDirs() error
}

type Download struct {
	Exit   chan struct{}
	UI     UI
	Config config.Config
	Env    Env
}

func (d *Download) Cmd() *cobra.Command {
	return &cobra.Command{
		Use:  "download",
		RunE: d.RunE,
	}
}

func (d *Download) RunE(cmd *cobra.Command, args []string) error {
	go func() {
		<-d.Exit
		os.Exit(128)
	}()

	if err := d.Env.CreateDirs(); err != nil {
		return errors.SafeWrap(err, "setup for download")
	}

	d.UI.Say("Downloading Resources...")
	return CacheSync(d.Config.Dependencies, d.Config.CacheDir, d.UI.Writer())
}

func CacheSync(dependencies resource.Catalog, cacheDir string, writer io.Writer) error {
	skipVerify := strings.ToLower(os.Getenv("CFDEV_SKIP_ASSET_CHECK"))

	cache := resource.Cache{
		Dir:                   cacheDir,
		HttpDo:                http.DefaultClient.Do,
		SkipAssetVerification: skipVerify == "true",
		Progress:              progress.New(writer),
		RetryWait:             time.Second,
		Writer:                writer,
	}

	if err := cache.Sync(dependencies); err != nil {
		return errors.SafeWrap(err, "Unable to sync assets")
	}
	return nil
}
