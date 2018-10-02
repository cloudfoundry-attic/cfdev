package profiler

type SystemProfiler struct {
}

func (*SystemProfiler)GetAvailableMemory() (int, error){
	return 0, nil
}