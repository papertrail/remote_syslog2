// +build !windows

package utils

import (
	"fmt"
	"io"
	"os"

	"github.com/VividCortex/godaemon"
	"github.com/nightlyone/lockfile"
)

const CanDaemonize = true

func Daemonize(logFilePath, pidFilePath string) {
	logFile, err := os.OpenFile(logFilePath, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0666)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Could not open local log file: %v", err)
		os.Exit(1)
	}

	stdout, stderr, err := godaemon.MakeDaemon(&godaemon.DaemonAttr{CaptureOutput: true})
	if err != nil {
		fmt.Fprintln(os.Stderr, "Could not Daemonize: %v", err)
		os.Exit(1)
	}

	go func() {
		io.Copy(logFile, stdout)
	}()
	go func() {
		io.Copy(logFile, stderr)
	}()

	lock, err := lockfile.New(pidFilePath)
	err = lock.TryLock()
	if err != nil {
		fmt.Println("Cannot lock \"%v\": %v", lock, err)
		os.Exit(1)
	}

}
