package main

import (
	"os"
	"path"
	"regexp"
	"time"

	"github.com/ActiveState/tail"
	"github.com/papertrail/remote_syslog2/syslog"
)

// Tails a single file
func tailone(
	file string,
	exclusions []*regexp.Regexp,
	severity syslog.Priority,
	facility syslog.Priority,
	poll bool,
	logger syslog.Logger,
	wr *WorkerRegistry,
) {
	defer wr.Remove(file)
	wr.Add(file)
	conf := tail.Config{
		ReOpen:    true,
		Follow:    true,
		MustExist: true,
		Poll:      poll,
		Location:  &tail.SeekInfo{0, os.SEEK_END},
	}
	t, err := tail.TailFile(file, conf)
	if err != nil {
		log.Errorf("%s", err)
		return
	}
	for line := range t.Lines {
		if !match(line.Text, exclusions) {
			log.Tracef("Forwarding: %s", line.Text)
			logger.Packets() <- &syslog.Packet{
				Severity: severity,
				Facility: facility,
				Time:     time.Now(),
				Hostname: logger.Hostname(),
				Tag:      path.Base(file),
				Message:  line.Text,
			}
		} else {
			log.Tracef("Not Forwarding: %s", line.Text)
		}
	}
	log.Errorf("Tail worker executed abnormally")
}
