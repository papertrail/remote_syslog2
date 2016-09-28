package main

import (
	"io"
	"net"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strconv"
	"time"

	"github.com/howbazaar/loggo"
	"github.com/hpcloud/tail"
	"github.com/papertrail/remote_syslog2/syslog"
	"github.com/papertrail/remote_syslog2/utils"
)

var log = loggo.GetLogger("")

type Server struct {
	config   *Config
	logger   *syslog.Logger
	registry *WorkerRegistry
}

func NewServer(config *Config) *Server {
	return &Server{
		config:   config,
		registry: NewWorkerRegistry(),
	}
}

func (s *Server) Start() error {
	if err := s.config.Validate(); err != nil {
		return err
	}

	if !s.config.NoDetach {
		utils.Daemonize(s.config.DebugLogFile, s.config.PidFile)
	}

	loggo.ConfigureLoggers(s.config.LogLevels)

	raddr := net.JoinHostPort(s.config.Destination.Host, strconv.Itoa(s.config.Destination.Port))
	log.Infof("Connecting to %s over %s", raddr, s.config.Destination.Protocol)

	var err error
	s.logger, err = syslog.Dial(
		s.config.Hostname,
		s.config.Destination.Protocol,
		raddr, s.config.RootCAs,
		s.config.ConnectTimeout,
		s.config.WriteTimeout,
		s.config.TcpMaxLineLength,
	)
	if err != nil {
		log.Errorf("Cannot connect to server: %v", err)
	}

	go s.tailFiles()

	for err = range s.logger.Errors {
		log.Errorf("Syslog error: %v", err)
	}

	return nil
}

// Tails a single file
func (s *Server) tailOne(file, tag string, whence int) {
	defer s.registry.Remove(file)
	s.registry.Add(file)

	t, err := tail.TailFile(file, tail.Config{
		ReOpen:    true,
		Follow:    true,
		MustExist: true,
		Poll:      s.config.Poll,
		Location:  &tail.SeekInfo{0, whence},
	})

	if err != nil {
		log.Errorf("%s", err)
		return
	}

	if tag == "" {
		tag = path.Base(file)
	}

	for line := range t.Lines {
		if !matchExps(line.Text, s.config.ExcludePatterns) {
			s.logger.Packets <- syslog.Packet{
				Severity: s.config.Severity,
				Facility: s.config.Facility,
				Time:     time.Now(),
				Hostname: s.logger.ClientHostname,
				Tag:      tag,
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
func (s *Server) tailFiles() {
	log.Debugf("Evaluating globs every %s", s.config.NewFileCheckInterval)
	firstPass := true

	for {
		s.globFiles(firstPass)
		time.Sleep(s.config.NewFileCheckInterval)
		firstPass = false
	}
}

//
func (s *Server) globFiles(firstPass bool) {
	log.Debugf("Evaluating file globs")
	for _, glob := range s.config.Files {

		tag := glob.Tag
		files, err := filepath.Glob(utils.ResolvePath(glob.Path))

		if err != nil {
			log.Errorf("Failed to glob %s: %s", glob.Path, err)
		} else if files == nil && firstPass {
			log.Errorf("Cannot forward %s, it may not exist", glob.Path)
		}

		for _, file := range files {
			switch {
			case s.registry.Exists(file):
				log.Debugf("Skipping %s because it is already running", file)
			case matchExps(file, s.config.ExcludeFiles):
				log.Debugf("Skipping %s because it is excluded by regular expression", file)
			default:
				log.Infof("Forwarding %s", file)

				whence := io.SeekStart

				// don't read the entire file on startup
				if firstPass {
					whence = io.SeekEnd
				}

				go s.tailOne(file, tag, whence)
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
	c, err := NewConfigFromEnv()
	if err != nil {
		log.Criticalf("Failed to configure the application: %s", err)
		os.Exit(1)
	}

	utils.AddSignalHandlers()

	s := NewServer(c)
	s.Start()
}
