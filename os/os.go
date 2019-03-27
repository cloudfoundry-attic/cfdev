package os

import "github.com/cloudfoundry/gosigar"

const bytesInMegabyte = 1048576

type Stats struct {
	AvailableMemory uint64
	TotalMemory     uint64
}

type OS struct{}

func (o *OS) Stats() (Stats, error) {
	mem := &sigar.Mem{}
	if err := mem.Get(); err != nil {
		return Stats{}, err
	}

	return Stats{
		AvailableMemory: mem.ActualFree / bytesInMegabyte,
		TotalMemory:     mem.Total / bytesInMegabyte,
	}, nil
}
