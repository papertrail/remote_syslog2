package main

import (
	"flag"
	"fmt"
	"github.com/sevenscale/remote_syslog2/papertrail"
	"github.com/sevenscale/remote_syslog2/syslog/certs"
	"io/ioutil"
	"launchpad.net/goyaml"
	"os"
	"time"
)

const (
	MinimumRefreshInterval = (time.Duration(10) * time.Second)
)

type ConfigFile struct {
	Files       []string
	Destination struct {
		Host     string
		Port     int
		Protocol string
	}
	Hostname string
	CABundle string `yaml:"ca_bundle"`
	//SetYAML is only called on pointers
	RefreshInterval *RefreshInterval `yaml:"refresh"`
}

type ConfigManager struct {
	Config ConfigFile
	Flags  struct {
		Hostname        string
		ConfigFile      string
		LogLevels       string
		RefreshInterval RefreshInterval
	}
	CertBundle certs.CertBundle
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
	cm.parseFlags()

	err := cm.readConfig()
	if err != nil {
		return err
	}

	err = cm.loadCABundle()
	if err != nil {
		return err
	}
	return nil
}

func (cm *ConfigManager) parseFlags() {
	flag.StringVar(&cm.Flags.ConfigFile, "config", "/etc/remote_syslog2/config.yaml", "the configuration file")
	flag.StringVar(&cm.Flags.Hostname, "hostname", "", "the name of this host")
	flag.StringVar(&cm.Flags.LogLevels, "log", "<root>=INFO", "\"logging configuration <root>=INFO;first=TRACE\"")
	flag.Var(&cm.Flags.RefreshInterval, "refresh", "How often to check for new files")
	flag.Parse()
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
	if err != nil {
		return fmt.Errorf("Could not read the config file: %s", err)
	}

	err = goyaml.Unmarshal(file, &cm.Config)
	if err != nil {
		return fmt.Errorf("Could not parse the config file: %s", err)
	}
	return nil
}

func (cm *ConfigManager) loadCABundle() error {
	bundle := certs.NewCertBundle()
	if cm.Config.CABundle == "" {
		log.Infof("Loading default certificates")

		loaded, err := bundle.LoadDefaultBundle()
		if loaded != "" {
			log.Infof("Loaded certificates from %s", loaded)
		}
		if err != nil {
			return err
		}

		log.Infof("Loading papertrail certificates")
		err = bundle.ImportBytes(papertrail.BundleCert())
		if err != nil {
			return err
		}

	} else {
		log.Infof("Loading certificates from %s", cm.Config.CABundle)
		err := bundle.ImportFromFile(cm.Config.CABundle)
		if err != nil {
			return err
		}

	}
	cm.CertBundle = bundle
	return nil
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

func (cm *ConfigManager) DestHost() string {
	return cm.Config.Destination.Host
}

func (cm ConfigManager) DestPort() int {
	return cm.Config.Destination.Port
}

func (cm *ConfigManager) DestProtocol() string {
	return cm.Config.Destination.Protocol
}

func (cm *ConfigManager) Files() []string {
	return cm.Config.Files
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
