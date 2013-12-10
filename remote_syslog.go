package main

import (
	"fmt"
	"github.com/ActiveState/tail"
	"github.com/howbazaar/loggo"
	"github.com/sevenscale/remote_syslog2/syslog"
	"os"
	"path"
	"path/filepath"
	"regexp"
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
func tailFiles(globs []string, excludedFiles []*regexp.Regexp, interval RefreshInterval, logger *syslog.Conn) {
	wr := NewWorkerRegistry()
	log.Debugf("Evaluating globs every %s", interval.Duration)
	logMissingFiles := true
	for {
		globFiles(globs, excludedFiles, logger, &wr, logMissingFiles)
		time.Sleep(interval.Duration)
		logMissingFiles = false
	}
}

//
func globFiles(globs []string, excludedFiles []*regexp.Regexp, logger *syslog.Conn, wr *WorkerRegistry, logMissingFiles bool) {
	log.Debugf("Evaluating file globs")
	for _, glob := range globs {

		files, err := filepath.Glob(glob)

		if err != nil {
			log.Errorf("Failed to glob %s: %s", glob, err)
		} else if files == nil && logMissingFiles {
			log.Errorf("Cannot forward %s, it may not exist", glob)
		}

		for _, file := range files {
			switch {
			case wr.Exists(file):
				log.Debugf("Skipping %s because it is already running", file)
			case matchExps(file, excludedFiles):
				log.Debugf("Skipping %s because it is excluded by regular expression", file)
			default:
				log.Infof("Forwarding %s", file)
				go tailOne(file, logger, wr)
			}
		}
	}
}

// Evaluates each regex against the string. If any one is a match
// the function returns true, otherwise it returns false
func matchExps(value string, expressions []*regexp.Regexp) bool {
	for _, exp := range expressions {
		if exp.MatchString(value) {
			return true
		}
	}
	return false
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
	go tailFiles(cm.Files(), cm.ExcludeFiles(), cm.RefreshInterval(), logger)

	ch := make(chan bool)
	<-ch
}
