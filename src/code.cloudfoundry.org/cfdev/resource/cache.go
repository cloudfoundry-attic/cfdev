package resource

import (
	"crypto/md5"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
)

type Cache struct {
	Dir                   string
	DownloadFunc          func(url, path string) error
	SkipAssetVerification bool
}

func (c *Cache) Sync(clog *Catalog) error {
	results, err := c.scan(c.Dir, clog)

	if err != nil {
		return err
	}

	for _, result := range results {
		filename := filepath.Join(c.Dir, result.item.Name)

		switch result.state {
		case missing:
			if err := c.downloadAndVerify(result, filename); err != nil {
				return err
			}
		case corrupt:
			if err := os.Remove(filename); err != nil {
				return err
			}
			if err := c.downloadAndVerify(result, filename); err != nil {
				return err
			}
		case valid:
		case unknown:
			os.Remove(filename)
		default:
			panic("unsupported sync state")
		}
	}

	return nil
}

func (c *Cache) downloadAndVerify(res result, filename string) error {
	if err := c.DownloadFunc(res.item.URL, filename); err != nil {
		return err
	}

	state, err := c.verifyFile(filename, res.item.MD5)

	if err != nil {
		return err
	}

	if state == corrupt {
		return fmt.Errorf("download file does not match checksum: %s %s",
			res.item.Name, res.item.MD5)
	}

	return nil
}

func (c *Cache) scan(dir string, clog *Catalog) ([]result, error) {
	var results []result

	for _, item := range clog.Items {
		itemPath := filepath.Join(dir, item.Name)
		_, err := os.Stat(itemPath)

		fileMissing := os.IsNotExist(err)

		if err != nil && !fileMissing {
			return nil, err
		}

		result := result{
			item: item,
		}

		if fileMissing {
			result.state = missing
		} else if result.state, err = c.verifyFile(itemPath, item.MD5); err != nil {
			return nil, err
		}

		results = append(results, result)
	}

	files, err := ioutil.ReadDir(c.Dir)
	if err != nil {
		return nil, err
	}

DirIteration:
	for _, f := range files {
		for _, item := range clog.Items {
			if item.Name == f.Name() {
				continue DirIteration
			}
		}

		results = append(results, result{
			item:  Item{Name: f.Name()},
			state: unknown,
		})

	}

	return results, nil
}

func (c *Cache) verifyFile(file string, expectedMD5 string) (state, error) {
	if c.SkipAssetVerification {
		return valid, nil
	}

	md5Hash, err := MD5(file)
	if err != nil {
		return corrupt, err
	}
	if md5Hash != expectedMD5 {
		return corrupt, nil
	}

	return valid, nil
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

type state int

const (
	unknown state = iota
	missing
	corrupt
	valid
)

type result struct {
	item  Item
	state state
}
