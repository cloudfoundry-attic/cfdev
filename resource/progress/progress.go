package progress

import (
	"fmt"
	"io"
	"strings"
)

type Progress struct {
	current              uint64
	currentLastCompleted uint64
	total                uint64
	lastPercentage       int
	writer               io.Writer
}

func New(writer io.Writer) *Progress {
	return &Progress{writer: writer}
}

func (c *Progress) Start(total uint64) {
	c.lastPercentage = -1
	c.current = 0
	c.total = total
	fmt.Fprintf(c.writer, "\rProgress: |%-21s| 0%%", ">")
}

func (c *Progress) Write(p []byte) (int, error) {
	c.current += uint64(len(p))
	c.display()
	return len(p), nil
}

func (c *Progress) Add(add uint64) {
	c.current += add
	c.display()
}

func (c *Progress) SetLastCompleted() {
	c.currentLastCompleted = c.current
}

func (c *Progress) ResetCurrent() {
	c.current = c.currentLastCompleted
	c.lastPercentage = c.lastPercentage + 1 //increment in order to print during retries
}

func (c *Progress) End() {
	fmt.Fprintf(c.writer, "\r\n")
}

func (c *Progress) display() {
	if c.total == 0 {
		fmt.Fprintf(c.writer, "\rProgress: %d bytes", c.current)
		return
	}
	percentage := int(c.current * 1000 / c.total)
	if c.lastPercentage == percentage {
		return
	}
	c.lastPercentage = percentage

	fmt.Fprintf(c.writer,
		"\rProgress: |%-21s| %.1f%%",
		strings.Repeat("=", percentage/50)+">",
		float64(percentage)/10.0)
}
