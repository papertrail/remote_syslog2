// +build darwin dragonfly freebsd linux nacl netbsd openbsd solaris

package lockfile

import (
	"os"
	"syscall"
)

func isProcessAlive(p *os.Process) error {
	err := p.Signal(os.Signal(syscall.Signal(0)))
	if err == nil {
		return nil
	}
	errno, ok := err.(syscall.Errno)
	if !ok {
		return ErrDeadOwner
	}

	switch errno {
	case syscall.ESRCH:
		return ErrDeadOwner
	case syscall.EPERM:
		return nil
	default:
		return err
	}
}
