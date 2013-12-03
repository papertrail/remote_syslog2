package main

import (
	"fmt"
	"github.com/ActiveState/tail"
	"github.com/howbazaar/loggo"
	"github.com/sevenscale/remote_syslog2/syslog"
	"os"
	"path"
	"path/filepath"
	"time"
)

var log = loggo.GetLogger("")

// Tails a single file
func tailOne(file string, logger *syslog.Conn, wr *WorkerRegistry) {
	defer wr.Remove(file)
	wr.Add(file)
	tailConfig := tail.Config{ReOpen: true, Follow: true, MustExist: true, Location: &tail.SeekInfo{0, os.SEEK_END}}

	t, err := tail.TailFile(file, tailConfig)

	if err != nil {
		log.Errorf("%s", err)
		return
	}

	for line := range t.Lines {
		p := syslog.Packet{
			Severity: syslog.SevInfo,
			Facility: syslog.LogLocal1, // todo: customize this
			Time:     time.Now(),
			Hostname: logger.Hostname(),
			Tag:      path.Base(file),
			Message:  line.Text,
		}
		err = logger.WritePacket(p)
		if err != nil {
			log.Errorf("%s", err)
		}

	}

	log.Errorf("Tail worker executed abnormally")
}

// Tails files speficied in the globs and re-evaluates the globs
// at the specified interval
func tailFiles(globs []string, interval RefreshInterval, logger *syslog.Conn) {
	wr := NewWorkerRegistry()
	log.Debugf("Evaluating globs every %s", interval.Duration)
	for {
		globFiles(globs, logger, &wr)
		time.Sleep(interval.Duration)
	}
}

//
func globFiles(globs []string, logger *syslog.Conn, wr *WorkerRegistry) {
	log.Debugf("Evaluating file globs")
	for _, glob := range globs {

		files, err := filepath.Glob(glob)

		if err != nil {
			log.Errorf("Failed to glob %s: %s", glob, err)
		} else if files == nil {
			log.Errorf("Cannot forward %s, it may not exist", glob)
		}

		for _, file := range files {
			if wr.Exists(file) {
				log.Debugf("Skipping %s", file)
			} else {
				log.Infof("Forwarding %s", file)
				go tailOne(file, logger, wr)
			}
		}
	}
}

func main() {
	cm := NewConfigManager()
	loggo.ConfigureLoggers(cm.LogLevels())
	hostname := cm.Hostname()

	destination := fmt.Sprintf("%s:%d", cm.DestHost(), cm.DestPort())

	log.Infof("Connecting to %s over %s", destination, cm.DestProtocol())
	logger, err := syslog.Dial(cm.DestProtocol(), destination, hostname, &cm.CertBundle)

	if err != nil {
		log.Criticalf("Cannot connect to server: %v", err)
		os.Exit(1)
	}

	go tailFiles(cm.Files(), cm.RefreshInterval(), logger)

	ch := make(chan bool)
	<-ch
}
