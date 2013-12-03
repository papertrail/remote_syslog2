package main

import (
	"fmt"
	"github.com/ActiveState/tail"
	"github.com/howbazaar/loggo"
	"github.com/sevenscale/remote_syslog2/syslog"
	"os"
	"path"
	"time"
)

var log = loggo.GetLogger("")

func tailFile(file string, logger *syslog.Conn) {
	tailConfig := tail.Config{ReOpen: true, Follow: true, MustExist: false, Location: &tail.SeekInfo{0, os.SEEK_END}}
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

	for _, file := range cm.Files() {
		log.Infof("Forwarding %s", file)
		go tailFile(file, logger)
	}

	ch := make(chan bool)
	<-ch
}
