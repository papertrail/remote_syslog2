// Package godaemon runs a program as a Unix daemon.
package godaemon

// Copyright (c) 2013 VividCortex, Inc. All rights reserved.
// Please see the LICENSE file for applicable license terms.

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"syscall"
	"time"
)

// The name of the env var that indicates what stage of daemonization we're at.
const stageVar = "__DAEMON_STAGE"

// DaemonAttr describes the options that apply to daemonization
type DaemonAttr struct {
	CaptureOutput bool // whether to capture stdout/stderr
}

/*
MakeDaemon turns the process into a daemon. But given the lack of Go's
support for fork(), MakeDaemon() is forced to run the process all over again,
from the start. Hence, this should probably be your first call after main
begins, unless you understand the effects of calling from somewhere else.
Keep in mind that the PID changes after this function is called, given
that it only returns in the child; the parent will exit without returning.

Options are provided as a DaemonAttr structure. In particular, setting the
CaptureOutput member to true will make the function return two io.Reader
streams to read the process' standard output and standard error, respectively.
That's useful if you want to capture things you'd normally lose given the
lack of console output for a daemon. Some libraries can write error conditions
to standard error or make use of Go's log package, that defaults to standard
error too. Having these streams allows you to capture them as required. (Note
that this function takes no action whatsoever on any of the streams.)

NOTE: If you use them, make sure NOT to take one of these readers and write
the data back again to standard output/error, or you'll end up with a loop.

Daemonizing is a 3-stage process. In stage 0, the program increments the
magical environment variable and starts a copy of itself that's a session
leader, with its STDIN, STDOUT, and STDERR disconnected from any tty. It
then exits.

In stage 1, the (new copy of) the program starts another copy that's not
a session leader, and then exits.

In stage 2, the (new copy of) the program chdir's to /, then sets the umask
and reestablishes the original value for the environment variable.
*/
func MakeDaemon(attrs *DaemonAttr) (io.Reader, io.Reader, error) {
	stage, advanceStage, resetEnv := getStage()

	// getExecutablePath() is OS-specific.
	procName, err := GetExecutablePath()

	if err != nil {
		err = fmt.Errorf("can't determine full path to executable: %v", err)
		return nil, nil, err
	}

	// If getExecutablePath() returns "" but no error, determinating the
	// executable path is not implemented on the host OS, so daemonization
	// is not supported.
	if len(procName) == 0 {
		err = fmt.Errorf("can't determine full path to executable")
		return nil, nil, err
	}

	if stage == 0 || stage == 1 {
		// Descriptors 0, 1 and 2 are fixed in the "os" package. If we close
		// them, the process may choose to open something else there, with bad
		// consequences if some write to os.Stdout or os.Stderr follows (even
		// from Go's library itself, through the default log package). We thus
		// reserve these descriptors to avoid that.
		nullDev, err := os.OpenFile("/dev/null", 0, 0)

		if err != nil {
			fmt.Fprintf(os.Stderr, "unable to open /dev/null: ", err, "\n")
			os.Exit(1)
		}

		files := make([]*os.File, 3, 5)
		files[0] = nullDev // stdin
		// FDs should not be changed; we're using literals (as opposed to
		// constants) on purpose to discourage such a practice.

		if stage == 1 && attrs.CaptureOutput {
			files = files[0:5]

			// stdout: write at fd:1, read at fd:3
			if files[3], files[1], err = os.Pipe(); err != nil {
				fmt.Fprintf(os.Stderr, "unable to create stdout pipe: ", err, "\n")
				os.Exit(1)
			}

			// stderr: write at fd:2, read at fd:4
			if files[4], files[2], err = os.Pipe(); err != nil {
				fmt.Fprintf(os.Stderr, "unable to create stderr pipe: ", err, "\n")
				os.Exit(1)
			}
		} else {
			files[1], files[2] = nullDev, nullDev
		}

		advanceStage()
		dir, _ := os.Getwd()
		attrs := os.ProcAttr{Dir: dir, Env: os.Environ(), Files: files}

		if stage == 0 {
			sysattrs := syscall.SysProcAttr{Setsid: true}
			attrs.Sys = &sysattrs
		}

		proc, err := os.StartProcess(procName, os.Args, &attrs)

		if err != nil {
			fmt.Fprintf(os.Stderr, "can't create process %s\n", procName)
			os.Exit(1)
		}

		proc.Release()
		os.Exit(0)
	}

	os.Chdir("/")
	syscall.Umask(0)
	resetEnv()

	var stdout, stderr *os.File
	if attrs.CaptureOutput {
		stdout = os.NewFile(uintptr(3), "stdout")
		stderr = os.NewFile(uintptr(4), "stderr")
	}
	return stdout, stderr, nil
}

// Daemonize is equivalent to MakeDaemon(&DaemonAttr{}). It is kept only for
// backwards API compatibility, but it's usage is otherwise discouraged. Use
// MakeDaemon() instead. The child parameter, previously used to tell whether
// to reset the environment or not (see MakeDaemon()), is currently ignored.
// The environment is reset in all cases.
func Daemonize(child ...bool) {
	MakeDaemon(&DaemonAttr{})
}

// Returns the current stage in the "daemonization process", that's kept in
// an environment variable. The variable is instrumented with a digital
// signature, to avoid misbehavior if it was present in the user's
// environment. The original value is restored after the last stage, so that
// there's no final effect on the environment the application receives.
func getStage() (stage int, advanceStage func(), resetEnv func()) {
	var origValue string
	stage = 0

	daemonStage := os.Getenv(stageVar)
	stageTag := strings.SplitN(daemonStage, ":", 2)
	stageInfo := strings.SplitN(stageTag[0], "/", 3)

	if len(stageInfo) == 3 {
		stageStr, tm, check := stageInfo[0], stageInfo[1], stageInfo[2]

		hash := sha1.New()
		hash.Write([]byte(stageStr + "/" + tm + "/"))

		if check != hex.EncodeToString(hash.Sum([]byte{})) {
			// This whole chunk is original data
			origValue = daemonStage
		} else {
			stage, _ = strconv.Atoi(stageStr)

			if len(stageTag) == 2 {
				origValue = stageTag[1]
			}
		}
	} else {
		origValue = daemonStage
	}

	advanceStage = func() {
		base := fmt.Sprintf("%d/%09d/", stage+1, time.Now().Nanosecond())
		hash := sha1.New()
		hash.Write([]byte(base))

		tag := base + hex.EncodeToString(hash.Sum([]byte{}))

		if err := os.Setenv(stageVar, tag+":"+origValue); err != nil {
			fmt.Fprintf(os.Stderr, "can't set %s (stage %d)\n", stageVar, stage)
			os.Exit(1)
		}
	}

	resetEnv = func() {
		if err := os.Setenv(stageVar, origValue); err != nil {
			fmt.Fprintf(os.Stderr, "can't reset %s\n", stageVar)
			os.Exit(1)
		}
	}

	return stage, advanceStage, resetEnv
}
