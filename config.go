package main

import (
	"crypto/x509"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/papertrail/remote_syslog2/papertrail"
	"github.com/papertrail/remote_syslog2/utils"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

var (
	config  *viper.Viper
	Version string

	MinimumRefreshInterval = time.Second
)

const (
	envPrefix = "rsyslog2"
)

// The global Config object for remote_syslog2 server. "mapstructure" tags
// signify the config file key names.
type Config struct {
	ConfigFile           string   `mapstructure:"config_file"`
	ConnectTimeout       int      `mapstructure:"connect_timeout"`
	WriteTimeout         int      `mapstructure:"write_timeout"`
	NewFileCheckInterval int      `mapstructure:"new_file_check_interval"`
	ExcludeFiles         []string `mapstructure:"exclude_files"`
	ExcludePatterns      []string `mapstructure:"exclude_patterns"`
	LogLevels            string   `mapstructure:"log_levels"`
	DebugLogFile         string   `mapstructure:"debug_log_file"`
	PidFile              string   `mapstructure:"pid_file"`
	TcpMaxLineLength     int      `mapstructure:"tcp_max_line_length"`
	NoDaemonize          bool     `mapstructure:"no_daemonize"`
	TCP                  bool     `mapstructure:"tcp"`
	TLS                  bool     `mapstructure:"tls"`
	Files                []interface{}
	Hostname             string
	Severity             string
	Facility             string
	Poll                 bool
	Destination          struct {
		Host     string
		Port     int
		Protocol string
	}
}

type LogFile struct {
	Path string
	Tag  string
}

func init() {
	config = viper.New()
	config.SetEnvPrefix(envPrefix)

	// set defaults for configuration values that aren't provided by flags here:
	config.SetDefault("destination.protocol", "udp")
	config.SetDefault("tcp_max_line_length", 99990)
	config.SetDefault("debug_log_file", "/dev/null")

	// set available commandline flags here:
	pflag.StringP("configfile", "c", "/etc/log_files.yml", "Path to config")
	config.BindPFlag("config_file", pflag.Lookup("configfile"))

	pflag.StringP("dest-host", "d", "", "Destination syslog hostname or IP")
	config.BindPFlag("destination.host", pflag.Lookup("dest-host"))

	pflag.IntP("dest-port", "p", 514, "Destination syslog port")
	config.BindPFlag("destination.port", pflag.Lookup("dest-port"))

	pflag.StringP("facility", "f", "user", "Facility")
	config.BindPFlag("facility", pflag.Lookup("facility"))

	hostname, _ := os.Hostname()
	pflag.String("hostname", hostname, "Local hostname to send from")
	config.BindPFlag("hostname", pflag.Lookup("hostname"))

	pflag.String("pid-file", "", "Location of the PID file")
	config.BindPFlag("pid_file", pflag.Lookup("pid-file"))

	pflag.StringP("severity", "s", "notice", "Severity")
	config.BindPFlag("severity", pflag.Lookup("severity"))

	pflag.Bool("tcp", false, "Connect via TCP (no TLS)")
	config.BindPFlag("tcp", pflag.Lookup("tcp"))

	pflag.Bool("tls", false, "Connect via TCP with TLS")
	config.BindPFlag("tls", pflag.Lookup("tls"))

	pflag.Bool("poll", false, "Detect changes by polling instead of inotify")
	config.BindPFlag("poll", pflag.Lookup("poll"))

	pflag.Int("new-file-check-interval", 10, "How often to check for new files")
	config.BindPFlag("new_file_check_interval", pflag.Lookup("new-file-check-interval"))

	pflag.String("debug-log-cfg", "", "the debug log file; overridden by -D/--no-detach")
	config.BindPFlag("debug_log_file", pflag.Lookup("debug-log-cfg"))

	pflag.String("log", "<root>=INFO", "set loggo config, like: --log=\"<root>=DEBUG\"")
	config.BindPFlag("log_levels", pflag.Lookup("log"))

	// only present this flag to systems that can daemonize
	if utils.CanDaemonize {
		pflag.BoolP("no-detach", "D", false, "Don't daemonize and detach from the terminal; overrides --debug-log-cfg")
		config.BindPFlag("no_daemonize", pflag.Lookup("no-detach"))
	}

	// deprecated flags
	pflag.Bool("no-eventmachine-tail", false, "No action, provided for backwards compatibility")
	pflag.Bool("eventmachine-tail", false, "No action, provided for backwards compatibility")

	// bind env vars to config automatically
	config.AutomaticEnv()
}

func NewConfig() (*Config, error) {
	pflag.Parse()

	c := &Config{}

	// read in config file if it's there
	config.SetConfigFile(config.GetString("config_file"))
	if err := config.ReadInConfig(); err != nil {
		return nil, err
	}

	// override daemonize setting for platforms that don't support it
	if !utils.CanDaemonize {
		config.Set("no_daemonize", true)
	}

	// unmarshal the viper config into our own struct
	if err := config.Unmarshal(c); err != nil {
		return nil, err
	}

	// collect any extra args passed and add them to our file list
	for f := range pflag.Args() {
		c.Files = append(c.Files, interface{}(f))
	}

	return c, nil
}

func (c *Config) Daemonize() bool {
	return !c.NoDaemonize
}

func (c *Config) GetDestinationProtocol() string {
	if c.TLS {
		return "tls"
	}

	if c.TCP {
		return "tcp"
	}

	return c.Destination.Protocol
}

func (c *Config) RootCAs() *x509.CertPool {
	if c.GetDestinationProtocol() == "tls" && c.Destination.Host == "logs.papertrailapp.com" {
		return papertrail.RootCA()
	}

	return nil
}

func (c *Config) GetFiles() []LogFile {
	var files []LogFile
	for _, f := range c.Files {
		switch val := f.(type) {
		case string:
			lf := strings.Split(val, "=")
			if len(lf) == 2 {
				files = append(files, LogFile{Tag: lf[0], Path: lf[1]})
			} else {
				files = append(files, LogFile{Path: val})
			}
		case map[interface{}]interface{}:
			var (
				tag  string
				path string
			)

			tag, _ = val["tag"].(string)
			path, _ = val["path"].(string)

			if path != "" {
				files = append(files, LogFile{Tag: tag, Path: path})
			}
		}
	}

	return files
}

func (c *Config) GetPidFile() string {
	if c.PidFile != "" {
		return c.PidFile
	}

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
