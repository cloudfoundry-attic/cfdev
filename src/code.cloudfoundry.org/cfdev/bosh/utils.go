package bosh

import (
	"io"
	"os"
	"time"

	"github.com/cloudfoundry/bosh-cli/ui"
)

type Logger struct{}

func (l *Logger) Debug(tag, msg string, args ...interface{})            {}
func (l *Logger) DebugWithDetails(tag, msg string, args ...interface{}) {}
func (l *Logger) Info(tag, msg string, args ...interface{})             {}
func (l *Logger) Warn(tag, msg string, args ...interface{})             {}
func (l *Logger) Error(tag, msg string, args ...interface{})            {}
func (l *Logger) ErrorWithDetails(tag, msg string, args ...interface{}) {}
func (l *Logger) HandlePanic(tag string)                                {}
func (l *Logger) ToggleForcedDebug()                                    {}
func (l *Logger) Flush() error                                          { return nil }
func (l *Logger) FlushTimeout(time.Duration) error                      { return nil }

type TaskReporter struct{}

func (t *TaskReporter) TaskStarted(int)             {}
func (t *TaskReporter) TaskFinished(int, string)    {}
func (t *TaskReporter) TaskOutputChunk(int, []byte) {}

type ReadCloserProxy struct {
	reader io.ReadCloser
}

func (p ReadCloserProxy) Seek(offset int64, whence int) (int64, error) {
	seeker, ok := p.reader.(io.Seeker)
	if ok {
		return seeker.Seek(offset, whence)
	}

	return 0, nil
}
func (p *ReadCloserProxy) Read(bs []byte) (int, error) {
	return p.reader.Read(bs)
}

func (p *ReadCloserProxy) Close() error {
	return p.reader.Close()
}

type FileReporter struct{}

func (f *FileReporter) TrackUpload(_ int64, r io.ReadCloser) ui.ReadSeekCloser {
	return &ReadCloserProxy{reader: r}
}
func (f *FileReporter) TrackDownload(int64, io.Writer) io.Writer { return os.Stdout }
