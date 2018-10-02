package profiler

import "github.com/shirou/gopsutil/mem"

type SystemProfiler struct {
}

func (s *SystemProfiler)GetAvailableMemory() (uint64, error){
	virtualMemory, err := mem.VirtualMemory()
	if err != nil {
		return 0, err
	}

	return virtualMemory.Free/1000000, nil
}

func (s *SystemProfiler)GetTotalMemory()(uint64, error){
	virtualMemory, err := mem.VirtualMemory()
	if err != nil {
		return 0, err
	}

	return virtualMemory.Total/1000000, nil
}