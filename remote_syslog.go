package main

import (
	"net"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strconv"
	"time"

	"github.com/ActiveState/tail"
	"github.com/howbazaar/loggo"
	"github.com/papertrail/remote_syslog2/syslog"
	"github.com/papertrail/remote_syslog2/utils"
)

var log = loggo.GetLogger("")

// Tails a single file
func tailOne(file string, excludePatterns []*regexp.Regexp, logger *syslog.Logger, wr *WorkerRegistry, severity syslog.Priority, facility syslog.Priority, poll bool) {
	defer wr.Remove(file)
	wr.Add(file)
	tailConfig := tail.Config{ReOpen: true, Follow: true, MustExist: true, Poll: poll, Location: &tail.SeekInfo{0, os.SEEK_END}}

	t, err := tail.TailFile(file, tailConfig)

	if err != nil {
		log.Errorf("%s", err)
		return
	}

	for line := range t.Lines {
		if !matchExps(line.Text, excludePatterns) {
			logger.Packets <- syslog.Packet{
				Severity: severity,
				Facility: facility,
				Time:     time.Now(),
				Hostname: logger.ClientHostname,
				Tag:      path.Base(file),
				Message:  line.Text,
			}
			log.Tracef("Forwarding: %s", line.Text)
		} else {
			log.Tracef("Not Forwarding: %s", line.Text)
		}

	}

	log.Errorf("Tail worker executed abnormally")
}

// Tails files speficied in the globs and re-evaluates the globs
// at the specified interval
func tailFiles(
	cm *ConfigManager,
	logger *syslog.Logger,
	severity syslog.Priority,
	facility syslog.Priority,
) {
	wr := NewWorkerRegistry()
	log.Debugf("Evaluating globs every %s", cm.RefreshInterval())
	logMissingFiles := true
	for {
		globFiles(cm, logger, &wr, logMissingFiles, severity, facility)
		time.Sleep(time.Duration(cm.RefreshInterval()))
		logMissingFiles = false
	}
}

//
func globFiles(
	cm *ConfigManager,
	logger *syslog.Logger,
	wr *WorkerRegistry,
	logMissingFiles bool,
	severity syslog.Priority,
	facility syslog.Priority,
) {
	log.Debugf("Evaluating file globs")
	for _, glob := range cm.Files() {
		files, err := filepath.Glob(utils.ResolvePath(glob))
		if err != nil {
			log.Errorf("Failed to glob %s: %s", glob, err)
		} else if files == nil && logMissingFiles {
			log.Errorf("Cannot forward %s, it may not exist", glob)
		}
		for _, file := range files {
			switch {
			case wr.Exists(file):
				log.Debugf("Skipping %s because it is already running", file)
			case matchExps(file, cm.ExcludeFiles()):
				log.Debugf("Skipping %s because it is excluded by regular expression", file)
			default:
				log.Infof("Forwarding %s", file)
				go tailOne(file, cm.ExcludePatterns(), logger, wr, severity, facility, cm.Poll())
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
	cm, err := NewConfigManager()

	if err != nil {
		log.Criticalf("Cannot initialize config manager: %v", err)
		os.Exit(1)
	}

	if cm.Daemonize() {
		utils.Daemonize(cm.DebugLogFile(), cm.PidFile())
	}

	loggo.ConfigureLoggers(cm.LogLevels())

	host, err := cm.DestHost()
	if err != nil {
		log.Criticalf("Invalid destination host: %v", err)
		os.Exit(1)
	}
	raddr := net.JoinHostPort(host, strconv.Itoa(cm.DestPort()))
	log.Infof("Connecting to %s over %s", raddr, cm.DestProtocol())
	rootcas, err := cm.RootCAs()
	if err != nil {
		log.Criticalf("Invalid root CAs: %v", err)
		os.Exit(1)
	}
	logger, err := syslog.Dial(cm.Hostname(), cm.DestProtocol(), raddr, rootcas)
	if err != nil {
		log.Criticalf("Cannot connect to server: %v", err)
		os.Exit(1)
	}

	severity, err := cm.Severity()
	if err != nil {
		log.Criticalf("Invalid severity: %v", err)
		os.Exit(1)
	}

	facility, err := cm.Facility()
	if err != nil {
		log.Criticalf("Invalid facility: %v", err)
		os.Exit(1)
	}

	go tailFiles(cm, logger, severity, facility)

	for err = range logger.Errors {
		log.Errorf("Syslog error: %v", err)
	}
}
