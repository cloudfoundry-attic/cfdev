package main

import (
	"io"
	"os"

	"github.com/natefinch/npipe"
)

func main() {
	c := make(chan os.Signal, 1)
	conn, err := npipe.Dial(`\\.\pipe\cfdev-com`)
	if err != nil {
		panic(err)
	}
	go io.Copy(conn, os.Stdin)
	go func() {
		for _ = range c {
			// sig is a ^C, handle it
			conn.Write([]byte("\033c"))
		}
	}()
	_, err = io.Copy(os.Stdout, conn)
}
