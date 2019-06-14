// +build !windows

package utils

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/VividCortex/godaemon"
	"github.com/nightlyone/lockfile"
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
	err = lock.TryLock()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Cannot lock \"%v\": %v\n", lock, err)
		os.Exit(1)
	}

}
