package lockfile

import (
	"os"
	"reflect"
	"syscall"
)

func isProcessAlive(p *os.Process) error {
	// Extract handle value from the os.Process struct to avoid the need
	// of a second, manually opened process handle.
	value := reflect.ValueOf(p)
	// Dereference *os.Process to os.Process
	value = value.Elem()
	field := value.FieldByName("handle")

	handle := syscall.Handle(field.Uint())

	var code uint32
	err := syscall.GetExitCodeProcess(handle, &code)
	if err != nil {
		return err
	}

	// code will contain the exit code of the process or 259 (STILL_ALIVE)
	// if the process has not exited yet.
	if code == 259 {
		return nil
	}

	return ErrDeadOwner
}
