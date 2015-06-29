package main

import (
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"time"

	"github.com/ogier/pflag"
	"github.com/papertrail/remote_syslog2/papertrail"
	"github.com/papertrail/remote_syslog2/syslog"
	"github.com/papertrail/remote_syslog2/utils"
	"gopkg.in/yaml.v2"
)

const (
	MinimumRefreshInterval = RefreshInterval(10 * time.Second)
	DefaultConfigFile      = "/etc/log_files.yml"
)

type ConfigFile struct {
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

type ConfigManager struct {
	Config    ConfigFile
	FlagFiles []string
	Flags     struct {
		Hostname        string
		DestHost        string
		DestPort        int
		ConfigFile      string
		LogLevels       string
		DebugLogFile    string
		PidFile         string
		RefreshInterval RefreshInterval
		UseTCP          bool
		UseTLS          bool
		NoDaemonize     bool
		Severity        syslog.Priority
		Facility        syslog.Priority
		Poll            bool
	}
}

func NewConfigManager() (*ConfigManager, error) {
	cm := &ConfigManager{
		Config: ConfigFile{
			ExcludeFiles:    RegexCollection{},
			ExcludePatterns: RegexCollection{},
		},
	}
	if err := cm.parseFlags(); err != nil {
		return nil, err
	}
	if err := cm.readConfig(); err != nil {
		return nil, err
	}
	return cm, nil
}

func (cm *ConfigManager) parseFlags() error {
	pflag.StringVarP(&cm.Flags.ConfigFile, "configfile", "c", DefaultConfigFile, "Path to config")
	pflag.StringVarP(&cm.Flags.DestHost, "dest-host", "d", "", "Destination syslog hostname or IP")
	pflag.IntVarP(&cm.Flags.DestPort, "dest-port", "p", 0, "Destination syslog port")
	if utils.CanDaemonize {
		pflag.BoolVarP(&cm.Flags.NoDaemonize, "no-detach", "D", false, "Don't daemonize and detach from the terminal")
	} else {
		cm.Flags.NoDaemonize = true
	}
	// facility
	var s string
	pflag.StringVarP(&s, "facility", "f", "user", "Facility")
	facility, err := syslog.Facility(s)
	if err != nil {
		return fmt.Errorf("%s is not a designated facility", s)
	}
	cm.Flags.Facility = facility
	pflag.StringVar(&cm.Flags.Hostname, "hostname", "", "Local hostname to send from")
	pflag.StringVar(&cm.Flags.PidFile, "pid-file", "", "Location of the PID file")
	// severity
	pflag.StringVarP(&s, "severity", "s", "notice", "Severity")
	severity, err := syslog.Severity(s)
	if err != nil {
		return fmt.Errorf("Invalid severity: %s", s)
	}
	cm.Flags.Severity = severity
	// --strip-color
	pflag.BoolVar(&cm.Flags.UseTCP, "tcp", false, "Connect via TCP (no TLS)")
	pflag.BoolVar(&cm.Flags.UseTLS, "tls", false, "Connect via TCP with TLS")
	pflag.BoolVar(&cm.Flags.Poll, "poll", false, "Detect changes by polling instead of inotify")
	pflag.StringVar(&s, "new-file-check-interval", "10s", "How often to check for new files")
	if err := cm.Flags.RefreshInterval.Set(s); err != nil {
		return err
	}
	_ = pflag.Bool("no-eventmachine-tail", false, "No action, provided for backwards compatibility")
	_ = pflag.Bool("eventmachine-tail", false, "No action, provided for backwards compatibility")
	pflag.StringVar(&cm.Flags.DebugLogFile, "debug-log-cfg", "", "the debug log file")
	pflag.StringVar(&cm.Flags.LogLevels, "log", "<root>=INFO", "\"logging configuration <root>=INFO;first=TRACE\"")
	pflag.Parse()
	cm.FlagFiles = pflag.Args()
	return nil
}

func (cm *ConfigManager) readConfig() error {
	log.Infof("Reading configuration file %s", cm.Flags.ConfigFile)
	return cm.loadConfigFile()
}

func (cm *ConfigManager) loadConfigFile() error {
	file, err := ioutil.ReadFile(cm.Flags.ConfigFile)
	// don't error if the default config file isn't found
	if os.IsNotExist(err) && cm.Flags.ConfigFile == DefaultConfigFile {
		return nil
	}
	if err != nil {
		return fmt.Errorf("Could not read the config file: %s", err)
	}
	if err = yaml.Unmarshal(file, &cm.Config); err != nil {
		return fmt.Errorf("Could not parse the config file: %s", err)
	}
	return cm.validateConfig()
}

func (cm *ConfigManager) validateConfig() error {
	// destination host
	if cm.Flags.DestHost == "" &&
		cm.Config.Destination.Host == "" {
		return fmt.Errorf("No destination hostname specified")
	}
	return nil
}

func (cm *ConfigManager) Daemonize() bool {
	return !cm.Flags.NoDaemonize
}

func (cm *ConfigManager) Hostname() string {
	switch {
	case cm.Flags.Hostname != "":
		return cm.Flags.Hostname
	case cm.Config.Hostname != "":
		return cm.Config.Hostname
	default:
		hostname, _ := os.Hostname()
		return hostname
	}
}

func (cm *ConfigManager) RootCAs() *x509.CertPool {
	host := cm.DestHost()
	if cm.DestProtocol() == "tls" &&
		host == "logs.papertrailapp.com" {
		return papertrail.RootCA()
	}
	return nil
}

func (cm *ConfigManager) DestHost() string {
	if cm.Flags.DestHost != "" {
		return cm.Flags.DestHost
	}
	return cm.Config.Destination.Host
}

func (cm *ConfigManager) DestPort() int {
	switch {
	case cm.Flags.DestPort != 0:
		return cm.Flags.DestPort
	case cm.Config.Destination.Port != 0:
		return cm.Config.Destination.Port
	default:
		return 514
	}
}

func (cm *ConfigManager) DestProtocol() string {
	switch {
	case cm.Flags.UseTLS:
		return "tls"
	case cm.Flags.UseTCP:
		return "tcp"
	case cm.Config.Destination.Protocol != "":
		return cm.Config.Destination.Protocol
	default:
		return "udp"
	}
}

func (cm *ConfigManager) Severity() syslog.Priority {
	return cm.Flags.Severity
}

func (cm *ConfigManager) Facility() syslog.Priority {
	return cm.Flags.Facility
}

func (cm *ConfigManager) Poll() bool {
	return cm.Flags.Poll
}

func (cm *ConfigManager) Files() []string {
	return append(cm.FlagFiles, cm.Config.Files...)
}

func (cm *ConfigManager) DebugLogFile() string {
	switch {
	case cm.Flags.DebugLogFile != "":
		return cm.Flags.DebugLogFile
	default:
		return "/dev/null"
	}
}

func (cm *ConfigManager) defaultPidFile() string {
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

func (cm *ConfigManager) PidFile() string {
	switch {
	case cm.Flags.PidFile != "":
		return cm.Flags.PidFile
	default:
		return cm.defaultPidFile()
	}
}

func (cm *ConfigManager) LogLevels() string {
	return cm.Flags.LogLevels
}

func (cm *ConfigManager) RefreshInterval() RefreshInterval {
	if cm.Flags.RefreshInterval != 0 {
		return cm.Flags.RefreshInterval
	}
	if cm.Config.RefreshInterval != 0 {
		return cm.Config.RefreshInterval
	}
	return MinimumRefreshInterval
}

func (cm *ConfigManager) ExcludeFiles() []*regexp.Regexp {
	return cm.Config.ExcludeFiles
}

func (cm *ConfigManager) ExcludePatterns() []*regexp.Regexp {
	return cm.Config.ExcludePatterns
}
