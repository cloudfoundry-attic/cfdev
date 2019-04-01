package hyperkit

import (
	"fmt"
	"runtime"
)

func SafeKill(pidfile, name string) error {
	return fmt.Errorf("SafeKill not implemented for %s", runtime.GOOS)
}