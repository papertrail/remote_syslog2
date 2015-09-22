package main

import (
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	"github.com/ogier/pflag"
	"github.com/papertrail/remote_syslog2/papertrail"
	"github.com/papertrail/remote_syslog2/syslog"
	"github.com/papertrail/remote_syslog2/utils"
	"gopkg.in/yaml.v2"
)

const (
	MIN_REFRESH_INTERVAL = RefreshInterval(10 * time.Second)
	DEFAULT_CONFIG_FILE  = "/etc/log_files.yml"
)

type Config struct {
	ConfigFile      string
	Files           []string
	DestHost        string
	DestPort        int
	Protocol        string
	Hostname        string
	RefreshInterval RefreshInterval
	ExcludeFiles    RegexCollection
	ExcludePatterns RegexCollection
	LogLevels       string
	DebugLogFile    string
	PidFile         string
	UseTCP          bool
	UseTLS          bool
	Daemonize       bool
	Severity        syslog.Priority
	Facility        syslog.Priority
	Poll            bool
	RootCAs         *x509.CertPool
}

type configfile struct {
	Files       []string
	Destination struct {
		Host     string `yaml:"host"`
		Port     int    `yaml:"port"`
		Protocol string `yaml:"protocol"`
	}
	Hostname        string          `yaml:"hostname"`
	RefreshInterval RefreshInterval `yaml:"new_file_check_interval"`
	ExcludeFiles    RegexCollection `yaml:"exclude_files"`
	ExcludePatterns RegexCollection `yaml:"exclude_patterns"`
}

func NewConfig() (*Config, error) {
	self := &Config{
		ConfigFile:      DEFAULT_CONFIG_FILE,
		ExcludeFiles:    RegexCollection{},
		ExcludePatterns: RegexCollection{},
		DestHost:        "localhost",
	}
	if err := self.load(); err != nil {
		return nil, err
	}
	// parse flags override config file
	if err := self.override(); err != nil {
		return nil, err
	}
	// check settings and set defaults if needed
	if err := self.validate(); err != nil {
		return nil, err
	}
	return self, nil
}

func (self *Config) load() error {
	log.Infof("Reading configuration file %s", self.ConfigFile)
	file, err := ioutil.ReadFile(self.ConfigFile)
	// don't error if the default config file isn't found
	if os.IsNotExist(err) && self.ConfigFile == DEFAULT_CONFIG_FILE {
		return nil
	}
	if err != nil {
		return fmt.Errorf("Could not read the config file: %s", err)
	}
	config := &configfile{}
	if err = yaml.Unmarshal(file, &config); err != nil {
		return fmt.Errorf("Could not parse the config file: %s", err)
	}
	self.Files = config.Files
	self.DestHost = config.Destination.Host
	self.DestPort = config.Destination.Port
	self.Protocol = config.Destination.Protocol
	self.Hostname = config.Hostname
	self.RefreshInterval = config.RefreshInterval
	self.ExcludeFiles = config.ExcludeFiles
	self.ExcludePatterns = config.ExcludePatterns
	return nil
}

func (self *Config) override() error {
	configfile := ""
	pflag.StringVarP(&configfile, "configfile", "c", "", "Path to config")
	if configfile != "" {
		self.ConfigFile = configfile
	}
	desthost := ""
	pflag.StringVarP(&desthost, "dest-host", "d", "", "Destination syslog hostname or IP")
	if desthost != "" {
		self.DestHost = desthost
	}
	destport := 0
	pflag.IntVarP(&destport, "dest-port", "p", 0, "Destination syslog port")
	if destport != 0 {
		self.DestPort = destport
	}
	foreground := true
	if utils.CanDaemonize {
		pflag.BoolVarP(&foreground, "no-detach", "D", false, "Don't daemonize and detach from the terminal")
	}
	self.Daemonize = !foreground
	// facility
	var s string
	pflag.StringVarP(&s, "facility", "f", "user", "Facility")
	facility, err := syslog.Facility(s)
	if err != nil {
		return fmt.Errorf("%s is not a designated facility", s)
	}
	self.Facility = facility
	pflag.StringVar(&self.Hostname, "hostname", "", "Local hostname to send from")
	pflag.StringVar(&self.PidFile, "pid-file", "", "Location of the PID file")
	// severity
	pflag.StringVarP(&s, "severity", "s", "notice", "Severity")
	severity, err := syslog.Severity(s)
	if err != nil {
		return fmt.Errorf("Invalid severity: %s", s)
	}
	self.Severity = severity
	// --strip-color
	pflag.BoolVar(&self.UseTCP, "tcp", false, "Connect via TCP (no TLS)")
	pflag.BoolVar(&self.UseTLS, "tls", false, "Connect via TCP with TLS")
	pflag.BoolVar(&self.Poll, "poll", false, "Detect changes by polling instead of inotify")
	pflag.StringVar(&s, "new-file-check-interval", "", "How often to check for new files")
	if s != "" {
		if err := self.RefreshInterval.Set(s); err != nil {
			return err
		}
	}
	_ = pflag.Bool("no-eventmachine-tail", false, "No action, provided for backwards compatibility")
	_ = pflag.Bool("eventmachine-tail", false, "No action, provided for backwards compatibility")
	logfile := ""
	pflag.StringVar(&logfile, "debug-log-cfg", "", "the debug log file")
	if logfile != "" {
		self.DebugLogFile = logfile
	}
	pflag.StringVar(&self.LogLevels, "log", "<root>=INFO", "\"logging configuration <root>=INFO;first=TRACE\"")
	pflag.Parse()
	self.Files = append(self.Files, pflag.Args()...)
	return nil
}

func (self *Config) validate() error {
	// hostname
	if self.Hostname == "" {
		hostname, err := os.Hostname()
		if err != nil {
			return fmt.Errorf("Could not obtain host name: %v", err)
		}
		self.Hostname = hostname
	}
	// destination host
	if self.DestHost == "" {
		return fmt.Errorf("No destination hostname specified")
	}
	// destination port
	if self.DestPort == 0 {
		self.DestPort = 514
	}
	// protocol
	if self.Protocol == "" {
		switch {
		case self.UseTLS:
			self.Protocol = "tls"
		case self.UseTCP:
			self.Protocol = "tcp"
		default:
			self.Protocol = "udp"
		}
	}
	// root CAs
	if self.Protocol == "tls" &&
		self.DestHost == "logs.papertrailapp.com" {
		self.RootCAs = papertrail.RootCA()
	}
	// log file
	if self.DebugLogFile == "" {
		self.DebugLogFile = "/dev/null"
	}
	// refresh interval
	if self.RefreshInterval == 0 {
		self.RefreshInterval = MIN_REFRESH_INTERVAL
	}
	// pid file
	if self.PidFile == "" {
		self.PidFile = self.defaultPidFile()
	}
	return nil
}

func (self *Config) defaultPidFile() string {
	pidFiles := []string{
		"/var/run/remote_syslog.pid",
		os.Getenv("HOME") + "/run/remote_syslog.pid",
		os.Getenv("HOME") + "/tmp/remote_syslog.pid",
		os.Getenv("HOME") + "/remote_syslog.pid",
		os.TempDir() + "/remote_syslog.pid",
		os.Getenv("TMPDIR") + "/remote_syslog.pid",
	}
	for _, f := range pidFiles {
		dir := filepath.Dir(f)
		dirStat, err := os.Stat(dir)
		if err != nil || dirStat == nil || !dirStat.IsDir() {
			continue
		}
		fd, err := os.OpenFile(f, os.O_WRONLY|os.O_CREATE, 0644)
		if err != nil {
			continue
		}
		fd.Close()
		return f
	}
	return "/tmp/remote_syslog.pid"
}
