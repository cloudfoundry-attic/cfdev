package garden

import (
	"io"
	"time"

	"code.cloudfoundry.org/cfdev/bosh"
	"code.cloudfoundry.org/cfdev/singlelinewriter"
)

type UI interface {
	Say(message string, args ...interface{})
	Writer() io.Writer
}

func (g *Garden) ReportProgress(ui UI, deploymentName string) {
	go func() {
		start := time.Now()
		lineWriter := singlelinewriter.New(ui.Writer())
		lineWriter.Say("  Uploading Releases")
		config, err := g.FetchBOSHConfig()
		b, err := bosh.New(config)
		if err == nil {
			ch := b.VMProgress(deploymentName)
			for p := range ch {
				if p.Total > 0 {
					ui.Say("  Progress: %d of %d (%s)", p.Done, p.Total, p.Duration.Round(time.Second))
				} else {
					ui.Say("  Uploaded Releases: %d (%s)", p.Releases, p.Duration.Round(time.Second))
				}
			}
			lineWriter.Close()
			ui.Say("  Done (%s)", time.Now().Sub(start).Round(time.Second))
		}
	}()
}
