// +build !windows

package utils

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"syscall"

	"github.com/VividCortex/godaemon"
	"github.com/leonsodhi/lockfile"
)

const CanDaemonize = true

func ResolvePath(path string) string {

	if filepath.IsAbs(path) {
		return path
	}

	return filepath.Join(os.Getenv("__DAEMON_CWD"), path)
}

func Daemonize(logFilePath, pidFilePath string) {
	if os.Getenv("__DAEMON_CWD") == "" {
		cwd, err := os.Getwd()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Cannot determine working directory: %v\n", err)
			os.Exit(1)
		}
		os.Setenv("__DAEMON_CWD", cwd)
	}

	logFile, err := os.OpenFile(logFilePath, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0666)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Could not open local log file: %v\n", err)
		os.Exit(1)
	}

	stdout, stderr, err := godaemon.MakeDaemon(&godaemon.DaemonAttr{CaptureOutput: true})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Could not Daemonize: %v\n", err)
		os.Exit(1)
	}

	go func() {
		io.Copy(logFile, stdout)
	}()
	go func() {
		io.Copy(logFile, stderr)
	}()

	lock, err := lockfile.New(pidFilePath)

	removePidFile := func() {
		fmt.Fprintf(os.Stderr, "Removing %s\n", pidFilePath)
		lock.Unlock()
	}

	err = lock.TryLock()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Cannot lock \"%v\", error was %v\n", lock, err)

		// Temporary workaround to transition from existing daemon
		if err.Error() == "os: process already finished" {
			removePidFile()
		}

		os.Exit(1)
	}

	SignalHandlers[syscall.SIGINT] = removePidFile  // Terminate
	SignalHandlers[syscall.SIGTERM] = removePidFile // Terminate
	SignalHandlers[syscall.SIGQUIT] = removePidFile // Stop gracefully
}
