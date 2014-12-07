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
	"launchpad.net/goyaml"
)

const (
	MinimumRefreshInterval = (time.Duration(10) * time.Second)
	DefaultConfigFile      = "/etc/log_files.yml"
)

type ConfigFile struct {
	Files       []string
	Destination struct {
		Host     string `yaml:"host"`
		Port     int    `yaml:"port"`
		Protocol string `yaml:"protocol"`
	}
	Hostname string `yaml:"hostname"`
	//SetYAML is only called on pointers
	RefreshInterval *RefreshInterval `yaml:"new_file_check_interval"`
	ExcludeFiles    *RegexCollection `yaml:"exclude_files"`
	ExcludePatterns *RegexCollection `yaml:"exclude_patterns"`
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
		Severity        string
		Facility        string
		Poll            bool
	}
}

type RefreshInterval struct {
	Duration time.Duration
}

func (r *RefreshInterval) String() string {
	return fmt.Sprint(*r)
}

func (r *RefreshInterval) Set(value string) error {
	d, err := time.ParseDuration(value)

	if err != nil {
		return err
	}

	if d < MinimumRefreshInterval {
		return fmt.Errorf("refresh interval must be greater than %s", MinimumRefreshInterval)
	}
	r.Duration = d
	return nil
}

func (r *RefreshInterval) SetYAML(tag string, value interface{}) bool {
	err := r.Set(value.(string))
	if err != nil {
		return false
	}
	return true
}

type RegexCollection []*regexp.Regexp

func (r *RegexCollection) Set(value string) error {
	exp, err := regexp.Compile(value)
	if err != nil {
		return err
	}
	*r = append(*r, exp)
	return nil
}

func (r *RegexCollection) String() string {
	return fmt.Sprint(*r)
}

func (r *RegexCollection) SetYAML(tag string, value interface{}) bool {
	items, ok := value.([]interface{})

	if !ok {
		return false
	}

	for _, item := range items {
		s, ok := item.(string)

		if !ok {
			return false
		}

		err := r.Set(s)
		if err != nil {
			panic(fmt.Sprintf("Failed to compile regex expression \"%s\"", s))
		}
	}

	return true
}

func NewConfigManager() ConfigManager {
	cm := ConfigManager{}
	err := cm.Initialize()

	if err != nil {
		log.Criticalf("Failed to configure the application: %s", err)
		os.Exit(1)
	}

	return cm
}

func (cm *ConfigManager) Initialize() error {
	cm.Config.ExcludeFiles = &RegexCollection{}
	cm.Config.ExcludePatterns = &RegexCollection{}
	cm.parseFlags()

	err := cm.readConfig()
	if err != nil {
		return err
	}
	return nil
}

func (cm *ConfigManager) parseFlags() {
	pflag.StringVarP(&cm.Flags.ConfigFile, "configfile", "c", DefaultConfigFile, "Path to config")
	pflag.StringVarP(&cm.Flags.DestHost, "dest-host", "d", "", "Destination syslog hostname or IP")
	pflag.IntVarP(&cm.Flags.DestPort, "dest-port", "p", 0, "Destination syslog port")
	if utils.CanDaemonize {
		pflag.BoolVarP(&cm.Flags.NoDaemonize, "no-detach", "D", false, "Don't daemonize and detach from the terminal")
	} else {
		cm.Flags.NoDaemonize = true
	}
	pflag.StringVarP(&cm.Flags.Facility, "facility", "f", "user", "Facility")
	pflag.StringVar(&cm.Flags.Hostname, "hostname", "", "Local hostname to send from")
	pflag.StringVar(&cm.Flags.PidFile, "pid-file", "", "Location of the PID file")
	// --parse-syslog
	pflag.StringVarP(&cm.Flags.Severity, "severity", "s", "notice", "Severity")
	// --strip-color
	pflag.BoolVar(&cm.Flags.UseTCP, "tcp", false, "Connect via TCP (no TLS)")
	pflag.BoolVar(&cm.Flags.UseTLS, "tls", false, "Connect via TCP with TLS")
	pflag.BoolVar(&cm.Flags.Poll, "poll", false, "Detect changes by polling instead of inotify")
	pflag.Var(&cm.Flags.RefreshInterval, "new-file-check-interval", "How often to check for new files")
	_ = pflag.Bool("no-eventmachine-tail", false, "No action, provided for backwards compatibility")
	_ = pflag.Bool("eventmachine-tail", false, "No action, provided for backwards compatibility")
	pflag.StringVar(&cm.Flags.DebugLogFile, "debug-log-cfg", "", "the debug log file")
	pflag.StringVar(&cm.Flags.LogLevels, "log", "<root>=INFO", "\"logging configuration <root>=INFO;first=TRACE\"")
	pflag.Parse()
	cm.FlagFiles = pflag.Args()
}

func (cm *ConfigManager) readConfig() error {
	log.Infof("Reading configuration file %s", cm.Flags.ConfigFile)
	err := cm.loadConfigFile()
	if err != nil {
		log.Errorf("%s", err)
		return err
	}
	return nil
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

	err = goyaml.Unmarshal(file, &cm.Config)
	if err != nil {
		return fmt.Errorf("Could not parse the config file: %s", err)
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
	if cm.DestProtocol() == "tls" && cm.DestHost() == "logs.papertrailapp.com" {
		return papertrail.RootCA()
	} else {
		return nil
	}
}

func (cm *ConfigManager) DestHost() string {
	switch {
	case cm.Flags.DestHost != "":
		return cm.Flags.DestHost
	case cm.Config.Destination.Host == "":
		log.Criticalf("No destination hostname specified")
		os.Exit(1)
	}
	return cm.Config.Destination.Host
}

func (cm ConfigManager) DestPort() int {
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
	s, err := syslog.Severity(cm.Flags.Severity)
	if err != nil {
		log.Criticalf("%s is not a designated facility", cm.Flags.Severity)
		os.Exit(1)
	}
	return s
}

func (cm *ConfigManager) Facility() syslog.Priority {
	f, err := syslog.Facility(cm.Flags.Facility)
	if err != nil {
		log.Criticalf("%s is not a designated facility", cm.Flags.Facility)
		os.Exit(1)
	}
	return f
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
	switch {
	case cm.Config.RefreshInterval != nil && cm.Flags.RefreshInterval.Duration != 0:
		return cm.Flags.RefreshInterval
	case cm.Config.RefreshInterval != nil:
		return *cm.Config.RefreshInterval
	case cm.Flags.RefreshInterval.Duration != 0:
		return cm.Flags.RefreshInterval
	}
	return RefreshInterval{Duration: MinimumRefreshInterval}
}

func (cm *ConfigManager) ExcludeFiles() []*regexp.Regexp {
	return *cm.Config.ExcludeFiles
}

func (cm *ConfigManager) ExcludePatterns() []*regexp.Regexp {
	return *cm.Config.ExcludePatterns
}
