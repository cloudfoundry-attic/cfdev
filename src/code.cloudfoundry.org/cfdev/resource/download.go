package resource

import (
	"fmt"
	"io"
	"net/http"
	"os"
)

type Downloader struct {
}

func (d *Downloader) Start(url, path string) error {
	resp, err := http.Get(url)

	if err != nil {
		return err
	}

	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("resource server returned status code %d", resp.StatusCode)
	}

	file, err := os.Create(path)

	if err != nil {
		return err
	}

	defer file.Close()

	_, err = io.Copy(file, resp.Body)

	if err != nil {
		os.Remove(path)
		return err
	}

	return nil
}
