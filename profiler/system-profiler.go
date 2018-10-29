package profiler

import (
	"github.com/cloudfoundry/gosigar"
)

const bytesInMegabyte = 1048576

type SystemProfiler struct {
}

func (s *SystemProfiler) GetAvailableMemory() (uint64, error) {
	mem := &sigar.Mem{}
	if err := mem.Get(); err != nil {
		return 0, err
	}
	return mem.ActualFree / bytesInMegabyte, nil
}

func (s *SystemProfiler) GetTotalMemory() (uint64, error) {
	mem := &sigar.Mem{}
	if err := mem.Get(); err != nil {
		return 0, err
	}
	return mem.Total / bytesInMegabyte, nil
}
