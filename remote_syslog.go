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

func main() {
	cm, err := NewConfigManager()
	if err != nil {
		log.Criticalf("Cannot initialize config manager: %v", err)
		os.Exit(1)
	}
	// run in the background if asked to
	if cm.Daemonize() {
		utils.Daemonize(cm.DebugLogFile(), cm.PidFile())
	}
	// set up logging
	loggo.ConfigureLoggers(cm.LogLevels())
	// connect to remote syslog
	host := cm.DestHost()
	raddr := net.JoinHostPort(host, strconv.Itoa(cm.DestPort()))
	log.Infof("Connecting to %s over %s", raddr, cm.DestProtocol())
	rootcas := cm.RootCAs()
	logger, err := syslog.Dial(
		cm.Hostname(),
		cm.DestProtocol(),
		raddr,
		rootcas,
	)
	if err != nil {
		log.Criticalf("Cannot connect to server: %v", err)
		os.Exit(1)
	}
	// tail files
	wr := NewWorkerRegistry()
	log.Debugf("Evaluating globs every %s", cm.RefreshInterval())
	logMissingFiles := true
	for {
		globFiles(cm, logger, &wr, logMissingFiles)
		for err = range logger.Errors {
			log.Errorf("Syslog error: %v", err)
		}
		time.Sleep(time.Duration(cm.RefreshInterval()))
		logMissingFiles = false
	}
}

// Tails a single file
func tailOne(
	cm *ConfigManager,
	file string,
	logger *syslog.Logger,
	wr *WorkerRegistry,
) {
	defer wr.Remove(file)
	wr.Add(file)
	tailConfig := tail.Config{
		ReOpen:    true,
		Follow:    true,
		MustExist: true,
		Poll:      cm.Poll(),
		Location:  &tail.SeekInfo{0, os.SEEK_END},
	}
	t, err := tail.TailFile(file, tailConfig)
	if err != nil {
		log.Errorf("%s", err)
		return
	}
	for line := range t.Lines {
		if !matchExps(line.Text, cm.ExcludePatterns()) {
			logger.Packets <- syslog.Packet{
				Severity: cm.Severity(),
				Facility: cm.Facility(),
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
func tailFiles(cm *ConfigManager, logger *syslog.Logger) {
	wr := NewWorkerRegistry()
	log.Debugf("Evaluating globs every %s", cm.RefreshInterval())
	logMissingFiles := true
	for {
		globFiles(cm, logger, &wr, logMissingFiles)
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
				go tailOne(cm, file, logger, wr)
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
