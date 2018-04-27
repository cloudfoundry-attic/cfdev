package resource

import (
	"crypto/md5"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

type Progress interface {
	io.Writer
	Start(total uint64)
	Add(add uint64)
	End()
}

type Cache struct {
	Dir                   string
	HttpDo                func(req *http.Request) (*http.Response, error)
	Progress              Progress
	SkipAssetVerification bool
}

func (c *Cache) Sync(clog *Catalog) error {
	c.Progress.Start(c.total(clog))
	for _, item := range clog.Items {
		if err := c.download(&item); err != nil {
			return err
		}
	}
	c.Progress.End()
	return c.removeUnknown(clog)
}

func (c *Cache) total(clog *Catalog) uint64 {
	var total uint64 = 0
	for _, item := range clog.Items {
		if item.InUse {
			total += item.Size
		}
	}
	return total
}

func (c *Cache) download(item *Item) error {
	if !item.InUse {
		return nil
	}

	if match, err := c.checksumMatches(filepath.Join(c.Dir, item.Name), item.MD5); err != nil {
		return err
	} else if match {
		c.Progress.Add(item.Size)
		return nil
	}
	if strings.HasPrefix(item.URL, "file://") {
		return c.copyFile(item)
	}

	tmpPath := filepath.Join(c.Dir, item.Name+".tmp."+item.MD5)
	if err := c.downloadHTTP(item.URL, tmpPath); err != nil {
		return err
	}
	if m, err := MD5(tmpPath); err != nil {
		return err
	} else if m != item.MD5 {
		os.Remove(tmpPath)
		return fmt.Errorf("md5 did not match: %s: %s != %s", item.Name, m, item.MD5)
	}
	return os.Rename(tmpPath, filepath.Join(c.Dir, item.Name))
}

func (c *Cache) downloadHTTP(url, tmpPath string) error {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}
	if fi, err := os.Stat(tmpPath); err == nil {
		c.Progress.Add(uint64(fi.Size()))
		req.Header.Add("Range", fmt.Sprintf("bytes=%d-", fi.Size()))
	}
	out, err := os.OpenFile(tmpPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0755)
	if err != nil {
		return err
	}
	defer out.Close()

	resp, err := c.HttpDo(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		if _, err = io.Copy(out, io.TeeReader(resp.Body, c.Progress)); err != nil {
			return err
		}
	} else if resp.StatusCode == 416 {
		// Possibly full file already downloaded
	} else {
		return fmt.Errorf("http: %s", resp.Status)
	}
	return nil
}

func (c *Cache) removeUnknown(clog *Catalog) error {
	known := make(map[string]bool, 0)
	for _, item := range clog.Items {
		known[item.Name] = true
	}
	files, err := ioutil.ReadDir(c.Dir)
	if err != nil {
		return err
	}
	for _, fi := range files {
		if !known[fi.Name()] {
			if err := os.Remove(filepath.Join(c.Dir, fi.Name())); err != nil {
				return err
			}
		}
	}
	return nil
}

func (c *Cache) checksumMatches(path, md5 string) (bool, error) {
	if c.SkipAssetVerification {
		return fileExists(path)
	}
	m, err := MD5(path)
	if os.IsNotExist(err) {
		return false, nil
	} else if err != nil {
		return false, err
	}
	return m == md5, nil
}

func (c *Cache) copyFile(item *Item) error {
	source, err := os.Open(strings.Replace(item.URL, "file://", "", 1))
	if err != nil {
		return err
	}
	defer source.Close()
	out, err := os.OpenFile(filepath.Join(c.Dir, item.Name), os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0755)
	if err != nil {
		return err
	}
	defer out.Close()
	_, err = io.Copy(out, io.TeeReader(source, c.Progress))
	return err
}

func MD5(file string) (string, error) {
	f, err := os.Open(file)
	if err != nil {
		return "", err
	}
	defer f.Close()
	h := md5.New()

	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}

	return fmt.Sprintf("%x", h.Sum(nil)), nil
}

func fileExists(file string) (bool, error) {
	_, err := os.Stat(file)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		} else {
			return false, err
		}
	}
	return true, nil
}
