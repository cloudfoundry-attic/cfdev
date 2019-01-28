package bosh

import (
	"code.cloudfoundry.org/cfdev/config"
	"code.cloudfoundry.org/cfdev/errors"
	boshdir "github.com/cloudfoundry/bosh-cli/director"
	"github.com/onsi/ginkgo"
	"time"
)

var VMProgressInterval = 1 * time.Second

type Bosh struct {
	dir boshdir.Director
}

func New(cfg config.Config) (*Bosh, error) {
	var (
		f         = boshdir.NewFactory(&Logger{})
		envs      = envsMapping(cfg)
		host, _   = envs["BOSH_ENVIRONMENT"]
		client, _ = envs["BOSH_CLIENT"]
		secret, _ = envs["BOSH_CLIENT_SECRET"]
		caCert, _ = envs["BOSH_CA_CERT"]
	)

	dir, err := f.New(boshdir.FactoryConfig{
		Host:         host,
		Port:         25555,
		Client:       client,
		ClientSecret: secret,
		CACert:       caCert,
	}, &TaskReporter{}, &FileReporter{})
	if err != nil {
		return nil, errors.SafeWrap(err, "failed to connect to bosh director")
	}
	return NewWithDirector(dir), nil
}

func NewWithDirector(dir boshdir.Director) *Bosh {
	return &Bosh{dir: dir}
}

const (
	UploadingReleases = "uploading-releases"
	Deploying         = "deploying"
	RunningErrand     = "running-errand"
)

type VMProgress struct {
	State    string
	Releases int
	Total    int
	Done     int
	Duration time.Duration
}

func (b *Bosh) VMProgress(deploymentName string) chan VMProgress {
	start := time.Now()
	var dep boshdir.Deployment

	for {
		var err error
		dep, err = b.dir.FindDeployment(deploymentName)
		if err == nil {
			break
		}
	}

	ch := make(chan VMProgress, 1)
	total := 0
	go func() {
		defer ginkgo.GinkgoRecover()

		for {
			time.Sleep(VMProgressInterval)

			vmInfos, err := dep.VMInfos()
			if err != nil || len(vmInfos) == 0 {
				if total == 0 {
					rels, err := b.dir.Releases()
					if err == nil {
						ch <- VMProgress{Releases: len(rels), Duration: time.Now().Sub(start)}
					}
				}
				continue
			}

			total = len(vmInfos)
			numDone := 0
			for _, v := range vmInfos {
				if v.ProcessState == "running" && len(v.Processes) > 0 {
					numDone++
				}
			}

			ch <- VMProgress{Total: total, Done: numDone, Duration: time.Now().Sub(start)}

			if numDone >= len(vmInfos) {
				close(ch)
				return
			}
		}
	}()

	return ch
}

func (b *Bosh) GetVMProgress(start time.Time, deploymentName string, isErrand bool) VMProgress {
	if isErrand {
		return VMProgress{State: RunningErrand, Duration: time.Now().Sub(start)}
	}

	var dep boshdir.Deployment

	for {
		var err error
		dep, err = b.dir.FindDeployment(deploymentName)
		if err == nil {
			break
		}
	}

	vmInfos, err := dep.VMInfos()
	if err != nil || len(vmInfos) == 0 {
		rels, err := b.dir.Releases()
		if err == nil {
			return VMProgress{State: UploadingReleases, Releases: len(rels), Duration: time.Now().Sub(start)}
		}
	}

	total := len(vmInfos)
	numDone := 0
	for _, v := range vmInfos {
		if v.ProcessState == "running" && len(v.Processes) > 0 {
			numDone++
		}
	}

	return VMProgress{State: Deploying, Total: total, Done: numDone, Duration: time.Now().Sub(start)}
}
