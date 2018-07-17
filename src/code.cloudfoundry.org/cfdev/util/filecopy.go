package util

import (
	"io"
	"os"
)

func CopyFile(source, target string) error {
	fh, err := os.Open(source)

	if err != nil {
		return err
	}

	defer fh.Close()

	fh2, err := os.OpenFile(target, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}

	defer fh2.Close()

	_, err = io.Copy(fh2, fh)
	return err
}
