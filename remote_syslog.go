package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"github.com/ActiveState/tail"
	"github.com/sevenscale/remote_syslog2/papertrail"
	"github.com/sevenscale/remote_syslog2/syslog"
	"github.com/sevenscale/remote_syslog2/syslog/certs"
	"io/ioutil"
	"log"
	"os"
	"path"
	"time"
)

func tailFile(file string, logger *syslog.Conn) error {
	tailConfig := tail.Config{ReOpen: true, Follow: true, MustExist: false, Location: &tail.SeekInfo{0, os.SEEK_END}}
	t, err := tail.TailFile(file, tailConfig)

	if err != nil {
		log.Println(err)
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
	CABundle string `json:"ca_bundle"`
}

func (c ConfigFile) GetHostname() string {
	if configHostname != "" {
		return configHostname
	} else {
		return c.Hostname
	}

}

var configHostname string

func main() {
	configFile := flag.String("config", "/etc/remote_syslog2/config.json", "the configuration file")
	flag.StringVar(&configHostname, "hostname", "", "the name of this host")
	flag.Parse()

	log.Printf("Reading configuration file %s", configFile)
	file, err := ioutil.ReadFile(*configFile)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	var config ConfigFile
	err = json.Unmarshal(file, &config)
	if err != nil {
		fmt.Println("The configfile is invalid: ", err)
		os.Exit(1)
	}

	hostname := config.GetHostname()
	if hostname == "" {
		hostname, _ = os.Hostname()
	}

	destination := fmt.Sprintf("%s:%d", config.Destination.Host, config.Destination.Port)
	cabundle := certs.NewCertBundle()
	if config.CABundle == "" {
		cabundle.LoadDefaultBundle()
		cabundle.ImportBytes(papertrail.BundleCert())
	} else {
		cabundle.ImportFromFile(config.CABundle)
	}

	log.Printf("Connecting to %s over %s", destination, config.Destination.Protocol)
	logger, err := syslog.Dial(config.Destination.Protocol, destination, hostname, &cabundle)

	if err != nil {
		log.Fatalf("Cannot connect to server: %v", err)
	}

	for _, file := range config.Files {
		log.Printf("Forwarding %s", file)
		go tailFile(file, logger)
	}

	ch := make(chan bool)
	<-ch
}
