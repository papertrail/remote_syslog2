// +build !windows

package utils

import (
	"io/ioutil"
	"os"
	"os/signal"
	"runtime"
	"syscall"
)

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

func AddSignalHandlers() {
	sigChan := make(chan os.Signal, 1)
	go func() {
		for sig := range sigChan {
			go func(sig os.Signal) {
				switch sig {
				case syscall.SIGUSR1:
					dumpStacks()
				}
			}(sig)

		}
	}()
	signal.Notify(sigChan, syscall.SIGUSR1)
}
