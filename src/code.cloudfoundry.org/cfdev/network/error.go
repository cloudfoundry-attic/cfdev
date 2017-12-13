package network

import "errors"

var UnprivilegedError = errors.New("not running as root")
