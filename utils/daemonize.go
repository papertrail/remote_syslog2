// +build !windows

package utils

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

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

	if !isDaemonized() {
		cwd, err := os.Getwd()
		if err != nil {
			fmt.Fprintln(os.Stderr, "Cannot determine working directory: %v", err)
			os.Exit(1)
		}
		os.Setenv("__DAEMON_CWD", cwd)

		logFile, err := os.OpenFile(logFilePath, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0666)
		if err != nil {
			fmt.Fprintln(os.Stderr, "Could not open local log file: %v", err)
			os.Exit(1)
		}

		err = daemonize([]*os.File{nil, logFile, logFile})
		if err != nil {
			fmt.Fprintln(os.Stderr, "Could not Daemonize: %v", err)
			os.Exit(1)
		}
	}

	lock, err := lockfile.New(pidFilePath)
	err = lock.TryLock()
	if err != nil {
		fmt.Println("Cannot lock \"%v\": %v", lock, err)
		os.Exit(1)
	}

}

func daemonize(files []*os.File) error {
	err := daemonChild('1', files)
	if err != nil {
		return err
	}
	os.Exit(0)
	panic("os.Exit returned")
}

func isDaemonized() bool {
	switch os.Getenv("__DAEMON_STEP") {
	default:
		return false
	case "1":
		err := daemonChild('2', nil)
		if err != nil {
			os.Exit(1)
			panic("os.Exit returned")
		}
		os.Exit(0)
		panic("os.Exit returned")
	case "2":
		os.Unsetenv("__DAEMON_STEP")
	}

	return true
}

func daemonChild(step rune, files []*os.File) error {
	if os.Getuid() != os.Geteuid() || os.Getgid() != os.Getegid() {
		// Can't rely on os.Executable being safe to execute, and the
		// fallback to os.Args[0] certainly wouldn't be safe.
		return errors.New("unsafe to daemonize suid/sgid executables")
	}

	name, err := os.Executable()
	if err != nil {
		name = os.Args[0]
	}

	attr := os.ProcAttr{}

	if step == '2' {
		attr.Dir = string(os.PathSeparator)
	}

	if step == '1' {
		devNull, err := os.OpenFile(os.DevNull, os.O_RDWR, 0)
		if err != nil {
			return err
		}
		defer devNull.Close()
		attr.Files = []*os.File{devNull, devNull, devNull}
	} else {
		attr.Files = []*os.File{os.Stdin, os.Stdout, os.Stderr}
	}

	if len(files) > len(attr.Files) {
		temp := make([]*os.File, len(files))
		copy(temp, attr.Files)
		attr.Files = temp
	}
	for i := range files {
		if files[i] != nil {
			attr.Files[i] = files[i]
		}
	}

	if step == '1' {
		attr.Sys = setsidAttr()
	}

	err = os.Setenv("__DAEMON_STEP", string(step))
	if err != nil {
		return err
	}
	defer os.Unsetenv("__DAEMON_STEP")

	process, err := os.StartProcess(name, os.Args, &attr)
	if err != nil {
		return err
	}

	if step == '1' {
		_, err = process.Wait()
		if err != nil {
			return err
		}
	} else {
		process.Release()
	}

	return nil
}
