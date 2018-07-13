package singlelinewriter

import (
	"bufio"
	"fmt"
	"io"
)

type UI interface {
	io.WriteCloser
	Say(message string, args ...interface{})
}

type ui struct {
	w  io.Writer
	pr io.Reader
	pw io.WriteCloser
}

func (ui *ui) Write(p []byte) (n int, err error) {
	return ui.pw.Write(p)
}
func (ui *ui) Close() error {
	return ui.pw.Close()
}
func (ui *ui) Say(message string, args ...interface{}) {
	fmt.Fprintf(ui.pw, "%s\n", fmt.Sprintf(message, args...))
}

func New(w io.Writer) UI {
	ui := &ui{w: w}
	ui.w.Write([]byte("\033[s\033[7l"))
	ui.pr, ui.pw = io.Pipe()
	go func() {
		scanner := bufio.NewScanner(ui.pr)
		for scanner.Scan() {
			w.Write([]byte("\033[u\033[J"))
			w.Write([]byte(scanner.Text()))
		}
		w.Write([]byte("\033[u\033[J"))
	}()
	return ui
}
