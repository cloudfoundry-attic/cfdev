package resource

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

type TarOpts struct {
	Include       string
	IncludeFolder string
	Exclude       string
	FlattenFolder bool
	Dst           string
}

func Untar(src string, dstOpts []TarOpts) error {
	f, err := os.Open(src)
	if err != nil {
		return err
	}
	defer f.Close()

	gzr, err := gzip.NewReader(f)
	if err != nil {
		return err
	}
	defer gzr.Close()

	tr := tar.NewReader(gzr)

	for {
		header, err := tr.Next()

		switch {
		case err == io.EOF:
			return nil
		case err != nil:
			return err
		case header == nil:
			continue
		}

		err = copyCurrentFileIfMatch(tr, header, dstOpts)
		if err != nil {
			return err
		}
	}
}

func copyCurrentFileIfMatch(tr *tar.Reader, header *tar.Header, opts []TarOpts) error {
	for _, opt := range opts {
		if header.Typeflag == tar.TypeDir {
			break
		}
		target := ""

		dir, filename := filepath.Split(header.Name)
		if opt.IncludeFolder != "" && !strings.Contains(dir, opt.IncludeFolder) {
			continue
		} else if opt.IncludeFolder != "" && strings.Contains(dir, opt.IncludeFolder) {
			if !opt.FlattenFolder {
				if err := os.MkdirAll(filepath.Join(opt.Dst, dir), 0755); err != nil {
					return err
				}

				target = filepath.Join(opt.Dst, header.Name)
			} else {
				target = filepath.Join(opt.Dst, filename)
			}
		} else if opt.Include == filepath.Base(header.Name) {
			if !opt.FlattenFolder {
				target = filepath.Join(opt.Dst, header.Name)
			} else {
				target = filepath.Join(opt.Dst, filepath.Base(header.Name))
			}
		}

		if target != "" {
			f, err := os.OpenFile(target, os.O_TRUNC|os.O_CREATE|os.O_RDWR, os.FileMode(header.Mode))
			if err != nil {
				return err
			}
			if _, err := io.Copy(f, tr); err != nil {
				return err
			}
			f.Close()
			break
		}
	}
	return nil
}

func Tar(src string, writers ...io.Writer) error {
	if _, err := os.Stat(src); err != nil {
		return fmt.Errorf("Unable to tar files - %v", err.Error())
	}

	mw := io.MultiWriter(writers...)

	gzw := gzip.NewWriter(mw)
	defer gzw.Close()

	tw := tar.NewWriter(gzw)
	defer tw.Close()

	return filepath.Walk(src, func(file string, fi os.FileInfo, err error) error {

		if err != nil {
			return err
		}

		header, err := tar.FileInfoHeader(fi, fi.Name())
		if err != nil {
			return err
		}

		header.Name = strings.TrimPrefix(strings.Replace(file, src, "", -1), string(filepath.Separator))

		if err := tw.WriteHeader(header); err != nil {
			return err
		}

		if !fi.Mode().IsRegular() {
			return nil
		}

		f, err := os.Open(file)
		if err != nil {
			return err
		}

		if _, err := io.Copy(tw, f); err != nil {
			return err
		}

		f.Close()

		return nil
	})
}
