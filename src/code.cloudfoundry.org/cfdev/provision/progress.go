package provision

import (
	"io"
	"time"

	"fmt"

	"code.cloudfoundry.org/cfdev/bosh"
	"code.cloudfoundry.org/cfdev/errors"
	"code.cloudfoundry.org/cfdev/singlelinewriter"
)

type UI interface {
	Say(message string, args ...interface{})
	Writer() io.Writer
}

func (c *Controller) DeployServices(ui UI, services []Service) error {
	config, err := c.FetchBOSHConfig()
	if err != nil {
		return err
	}

	b, err := bosh.New(config)
	if err != nil {
		return err
	}

	errChan := make(chan error, 1)

	for _, service := range services {
		start := time.Now()
		ui.Say("Deploying %s...", service.Name)

		go func(handle string, script string) {
			errChan <- c.DeployService(handle, script)
		}(service.Handle, service.Script)

		err := c.report(start, ui, b, service, errChan)
		if err != nil {
			return err
		}
	}

	return nil
}

func (c *Controller) report(start time.Time, ui UI, b *bosh.Bosh, service Service, errChan chan error) error {
	for {
		select {
		case err := <-errChan:
			if err != nil {
				return errors.SafeWrap(err, fmt.Sprintf("Failed to deploy %s", service.Name))
			}

			ui.Writer().Write([]byte(fmt.Sprintf("\r\033[K  Done (%s)\n", time.Now().Sub(start).Round(time.Second))))
			return nil
		default:
			p := b.GetVMProgress(start, service.Deployment, service.IsErrand)

			switch p.State {
			case bosh.UploadingReleases:
				ui.Writer().Write([]byte(fmt.Sprintf("\r\033[K  Uploaded Releases: %d (%s)", p.Releases, p.Duration.Round(time.Second))))
			case bosh.Deploying:
				ui.Writer().Write([]byte(fmt.Sprintf("\r\033[K  Progress: %d of %d (%s)", p.Done, p.Total, p.Duration.Round(time.Second))))
			case bosh.RunningErrand:
				ui.Writer().Write([]byte(fmt.Sprintf("\r\033[K  Running errand (%s)", p.Duration.Round(time.Second))))
			}

			time.Sleep(time.Second)
		}
	}
}

func (c *Controller) ReportProgress(ui UI, deploymentName string) {
	go func() {
		start := time.Now()
		lineWriter := singlelinewriter.New(ui.Writer())
		lineWriter.Say("  Uploading Releases")
		config, err := c.FetchBOSHConfig()
		b, err := bosh.New(config)
		if err == nil {
			ch := b.VMProgress(deploymentName)
			for p := range ch {
				if p.Total > 0 {
					lineWriter.Say("  Progress: %d of %d (%s)", p.Done, p.Total, p.Duration.Round(time.Second))
				} else {
					lineWriter.Say("  Uploaded Releases: %d (%s)", p.Releases, p.Duration.Round(time.Second))
				}
			}
			lineWriter.Close()
			ui.Say("  Done (%s)", time.Now().Sub(start).Round(time.Second))
		}
	}()
}
