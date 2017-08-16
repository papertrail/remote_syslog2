package utils

import "syscall"

func setsidAttr() *syscall.SysProcAttr {
	CREATE_NO_WINDOW := uint32(0x08000000)
	return &syscall.SysProcAttr{CreationFlags: CREATE_NO_WINDOW}
}
