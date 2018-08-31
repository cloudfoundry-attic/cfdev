package unsetboshallproxy

import (
	"os"
)

func init() {
	os.Unsetenv("BOSH_ALL_PROXY")
}
