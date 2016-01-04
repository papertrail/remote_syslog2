// +build !windows

package utils

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/signal"
	"runtime"
	"syscall"
)

var SignalHandlers = map[os.Signal]func(){
	syscall.SIGUSR1: dumpStacks, // Dump goroutines stacks
}

func AddSignalHandlers() {
	var trapped []os.Signal
	for k := range SignalHandlers {
		trapped = append(trapped, k)
	}

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, trapped...)

	go func() {
		for sig := range signals {
			if f, found := SignalHandlers[sig]; found {
				fmt.Fprintf(os.Stderr, "Handling signal: %v\n", sig)
				f()
			}

			switch sig {
			case syscall.SIGINT, syscall.SIGTERM:
				os.Exit(128 + int(sig.(syscall.Signal)))
			case syscall.SIGQUIT:
				os.Exit(0)
			}
		}
	}()
}

// Mostly copied from Docker
func dumpStacks() {
	var (
		buf       []byte
		stackSize int
	)

	// Continually grab the trace until the buffer size exceeds its length
	bufferLen := 16384
	for stackSize == len(buf) {
		buf = make([]byte, bufferLen)
		stackSize = runtime.Stack(buf, true)
		bufferLen *= 2
	}
	buf = buf[:stackSize]

	f, err := ioutil.TempFile("", "r_s_stacktrace")
	defer f.Close()
	if err == nil {
		f.WriteString(string(buf))
	}

}
