package main

import (
	"errors"
	"flag"
	"fmt"
	"github.com/ActiveState/tail"
	"github.com/sevenscale/remote_syslog2/papertrail"
	"github.com/sevenscale/remote_syslog2/syslog"
	"github.com/sevenscale/remote_syslog2/syslog/certs"
	"github.com/shxsun/klog"
	"io/ioutil"
	"launchpad.net/goyaml"
	"os"
	"path"
	"time"
)

var log *klog.Logger

func init() {
	log = klog.NewLogger(nil, "")
	log.SetLevel(klog.LDebug)
}

func tailFile(file string, logger *syslog.Conn) error {
	tailConfig := tail.Config{ReOpen: true, Follow: true, MustExist: false, Location: &tail.SeekInfo{0, os.SEEK_END}}
	t, err := tail.TailFile(file, tailConfig)

	if err != nil {
		log.Error(err)
		return err
	}

	for line := range t.Lines {
		p := syslog.Packet{
			Severity: syslog.SevInfo,
			Facility: syslog.LogLocal1, // todo: customize this
			Time:     time.Now(),
			Hostname: logger.Hostname(),
			Tag:      path.Base(file),
			Message:  line.Text,
		}
		err = logger.WritePacket(p)
		if err != nil {
			return err
		}

	}

	return errors.New("Tail worker executed abnormally")
}

type ConfigFile struct {
	Files       []string
	Destination struct {
		Host     string
		Port     int
		Protocol string
	}
	Hostname string
	CABundle string `yaml:"ca_bundle"`
}

type ConfigManager struct {
	Config ConfigFile
	Flags  struct {
		Hostname   string
		ConfigFile string
		LogLevel   int
	}
	CertBundle certs.CertBundle
}

func NewConfigManager() ConfigManager {
	cm := ConfigManager{}
	err := cm.Initialize()

	if err != nil {
		log.Fatalf("Failed to configure the application: %s", err)
	}

	return cm
}

func (cm *ConfigManager) Initialize() error {
	cm.parseFlags()

	err := cm.readConfig()
	if err != nil {
		return err
	}

	err = cm.loadConfigFile()
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
	flag.IntVar(&cm.Flags.LogLevel, "loglevel", 4, "Log Level 0=Debug .. 4=Fatal")
	flag.Parse()
}

func (cm *ConfigManager) readConfig() error {
	log.Infof("Reading configuration file %s", cm.Flags.ConfigFile)
	err := cm.loadConfigFile()
	if err != nil {
		log.Error(err)
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
		log.Info("Loading default certificates")

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

func (cm *ConfigManager) LogLevel() klog.Level {
	switch cm.Flags.LogLevel {
	case 0:
		return klog.LDebug
	case 1:
		return klog.LInfo
	case 2:
		return klog.LWarning
	case 3:
		return klog.LError
	case 4:
		return klog.LFatal
	default:
		log.Errorf("Invalid logger level %d, assuming 1=Info", cm.Flags.LogLevel)
		return klog.LInfo
	}
}

func (cm *ConfigManager) Files() []string {
	return cm.Config.Files
}

func main() {
	cm := NewConfigManager()
	log.SetLevel(cm.LogLevel())
	hostname := cm.Hostname()

	destination := fmt.Sprintf("%s:%d", cm.DestHost(), cm.DestPort())

	log.Infof("Connecting to %s over %s", destination, cm.DestProtocol())
	logger, err := syslog.Dial(cm.DestProtocol(), destination, hostname, &cm.CertBundle)

	if err != nil {
		log.Fatalf("Cannot connect to server: %v", err)
	}

	for _, file := range cm.Files() {
		log.Infof("Forwarding %s", file)
		go tailFile(file, logger)
	}

	ch := make(chan bool)
	<-ch
}
