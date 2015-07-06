package main

import (
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
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
	// creating a new flag set allows for calling
	// this method multiple times, e.g. during testing
	flags := pflag.NewFlagSet(os.Args[0], pflag.ExitOnError)
	configfile := flags.StringP("configfile", "c", "", "Path to config")
	desthost := flags.StringP("dest-host", "d", "", "Destination syslog hostname or IP")
	destport := flags.IntP("dest-port", "p", 0, "Destination syslog port")
	logfile := flags.String("debug-log-cfg", "", "the debug log file")
	foreground := flags.BoolP("no-detach", "D", false, "Don't daemonize and detach from the terminal")
	facility := flags.StringP("facility", "f", "user", "Facility")
	severity := flags.StringP("severity", "s", "notice", "Severity")
	refresh := flags.String("new-file-check-interval", "", "How often to check for new files")
	//
	hostname := flags.String("hostname", "", "Local hostname to send from")
	pidfile := flags.String("pid-file", "", "Location of the PID file")
	// --strip-color
	poll := flags.Bool("poll", false, "Detect changes by polling instead of inotify")
	loglevels := flags.String("log", "<root>=INFO", "\"logging configuration <root>=INFO;first=TRACE\"")
	_ = flags.Bool("no-eventmachine-tail", false, "No action, provided for backwards compatibility")
	_ = flags.Bool("eventmachine-tail", false, "No action, provided for backwards compatibility")
	// work around the fact that we can't tell if a bool flag was set
	tcp := flags.Bool("tcp", false, "Connect via TCP (no TLS)")
	tls := flags.Bool("tls", true, "Connect via TCP with TLS")
	tcpExists := self.boolFlagExists("--tcp", os.Args)
	tlsExists := self.boolFlagExists("--tls", os.Args)
	// parse
	flags.Parse(os.Args[1:])
	// reload config file if needed
	if *configfile != "" {
		self.ConfigFile = *configfile
		if err := self.load(); err != nil {
			return err
		}
	}
	// set
	if utils.CanDaemonize {
		self.Daemonize = !*foreground
	}
	v, err := syslog.Facility(*facility)
	if err != nil {
		return fmt.Errorf("%s is not a designated facility", *facility)
	}
	self.Facility = v
	v, err = syslog.Severity(*severity)
	if err != nil {
		return fmt.Errorf("Invalid severity: %s", *severity)
	}
	self.Severity = v
	if *refresh != "" {
		err := yaml.Unmarshal([]byte(*refresh), &self.RefreshInterval)
		if err != nil {
			return err
		}
	}
	if *pidfile != "" {
		self.PidFile = *pidfile
	}
	self.Poll = *poll
	if *loglevels != "" {
		self.LogLevels = *loglevels
	}
	self.Files = append(self.Files, flags.Args()...)
	// override
	switch {
	case tlsExists && *tls == true:
		self.Protocol = "tls"
	case tcpExists && *tcp == true:
		self.Protocol = "tcp"
	case self.Protocol != "":
		// already set
	default:
		self.Protocol = "udp"
	}
	if *desthost != "" {
		self.DestHost = *desthost
	}
	if *destport != 0 {
		self.DestPort = *destport
	}
	if *logfile != "" {
		self.DebugLogFile = *logfile
	}
	if *hostname != "" {
		self.Hostname = *hostname
	}
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

func (self *Config) boolFlagExists(flag string, args []string) bool {
	for _, v := range args {
		if strings.HasPrefix(v, flag) {
			return true
		}
	}
	return false
}
