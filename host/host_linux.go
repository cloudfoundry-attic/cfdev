package host

func (*Host) CheckRequirements() error {
	return nil
}
func (h *Host) Version() (string, error) {
	return "", nil
}
