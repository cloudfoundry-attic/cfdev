package hypervisor

type VM struct {
	Name     string
	DepsIso  string
	MemoryMB int
	CPUs     int
}
