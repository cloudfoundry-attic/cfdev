package resource

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
)

type Downloader struct {
}

func (d *Downloader) Start(uri, path string) error {
	var src io.Reader
	parsed, err := url.Parse(uri)
	if err == nil && parsed.Scheme == "file" {
		fh, err := os.Open(parsed.Path)
		if err != nil {
			return err
		}
		defer fh.Close()
		src = fh
	} else {
		resp, err := http.Get(uri)
		if err != nil {
			return err
		}

		defer resp.Body.Close()

		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			return fmt.Errorf("resource server returned status code %d", resp.StatusCode)
		}
		src = resp.Body
	}

	file, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0777)

	if err != nil {
		return err
	}

	defer file.Close()

	_, err = io.Copy(file, src)

	if err != nil {
		os.Remove(path)
		return err
	}

	return nil
}
