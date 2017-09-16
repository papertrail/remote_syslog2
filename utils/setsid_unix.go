// +build !windows

package utils

import "syscall"

func setsidAttr() *syscall.SysProcAttr {
	return &syscall.SysProcAttr{Setsid: true}
}
