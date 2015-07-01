package main

import (
	"net"
	"os"
	"strconv"

	"github.com/howbazaar/loggo"
	"github.com/papertrail/remote_syslog2/syslog"
	"github.com/papertrail/remote_syslog2/utils"
)

var log = loggo.GetLogger("")

func main() {
	config, err := NewConfig()
	if err != nil {
		log.Criticalf("Cannot initialize config manager: %v", err)
		os.Exit(1)
	}
	// run in the background if asked to
	if config.Daemonize {
		utils.Daemonize(config.DebugLogFile, config.PidFile)
	}
	// set up logging
	loggo.ConfigureLoggers(config.LogLevels)
	// connect to remote syslog
	host := config.DestHost
	raddr := net.JoinHostPort(host, strconv.Itoa(config.DestPort))
	log.Infof("Connecting to %s over %s", raddr, config.Protocol)
	rootcas := config.RootCAs
	logger, err := syslog.Dial(
		config.Hostname,
		config.Protocol,
		raddr,
		rootcas,
	)
	if err != nil {
		log.Criticalf("Cannot connect to server: %v", err)
		os.Exit(1)
	}
	// tail files
	wr := NewWorkerRegistry()
	log.Debugf("Evaluating globs every %s", config.RefreshInterval)
	warn := true
	go tailFiles(config, logger, wr, warn)
	for err = range logger.Errors() {
		log.Errorf("Syslog error: %v", err)
	}
}
