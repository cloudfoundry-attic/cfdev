package profiler

import "github.com/shirou/gopsutil/mem"

type SystemProfiler struct {
}

func (*SystemProfiler)GetAvailableMemory() (uint64, error){
	virtualMemory, err := mem.VirtualMemory()
	if err != nil {
		return 0, nil
	}

	return virtualMemory.Free/1000000, nil
}