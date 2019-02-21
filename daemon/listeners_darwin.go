package daemon

// #include <stdlib.h>
// int launch_activate_socket(const char *name, int **fds, size_t *cnt);
import "C"

import (
	"fmt"
	"net"
	"os"
	"unsafe"
)

func Listeners(name string) ([]net.Listener, error) {
	files, err := files(name)
	if err != nil {
		return nil, err
	}

	listeners := make([]net.Listener, len(files), len(files))
	for i, file := range files {
		if listener, err := net.FileListener(file); err != nil {
			return nil, err
		} else {
			listeners[i] = listener
		}
	}

	return listeners, nil
}

func files(name string) ([]*os.File, error) {
	var fds *C.int
	cnt := C.size_t(0)

	err := C.launch_activate_socket(C.CString(name), &fds, &cnt)
	if err != 0 {
		return nil, fmt.Errorf("error activating launchd socket %s: %d", name, err)
	}

	ptr := unsafe.Pointer(fds)
	defer C.free(ptr)
	arr := (*[10]C.int)(ptr)

	files := make([]*os.File, int(cnt), int(cnt))
	for i := 0; i < int(cnt) && i < 10; i++ {
		files[i] = os.NewFile(uintptr(int(arr[i])), "")
	}

	return files, nil
}
