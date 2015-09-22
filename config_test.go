package main

import (
	"os"
	"strings"
	"time"

	"github.com/papertrail/remote_syslog2/syslog"
	. "gopkg.in/check.v1"
)

type ConfigSuite struct {
}

var _ = Suite(&ConfigSuite{})

func (s *ConfigSuite) SetUpSuite(c *C) {
}

func (s *ConfigSuite) TestOverride(c *C) {
	config := &Config{
		ConfigFile:      "test/config_with_host.yaml",
		ExcludeFiles:    RegexCollection{},
		ExcludePatterns: RegexCollection{},
		DestHost:        "localhost",
	}
	c.Assert(config.load(), IsNil)
	c.Assert(config.Files, HasLen, 1)
	c.Assert(config.Files[0], Equals, "locallog.txt")
	c.Assert(config.DestHost, Equals, "logs.papertrailapp.com")
	c.Assert(config.DestPort, Equals, 514)
	c.Assert(config.Protocol, Equals, "tls")
	c.Assert(config.Hostname, Equals, "test-host-from-config")
	c.Assert(config.RefreshInterval, Equals, RefreshInterval(30*time.Second))
	c.Assert(config.validate(), IsNil)
	// fudge the args
	os.Args = os.Args[0:1]
	args := `-d foo -p 1111 --debug-log-cfg=xxx --new-file-check-interval=20s --hostname=yyy --pid-file=pidfile --tcp --tls --poll --log=<root>=DEBUG a b`
	a := strings.Split(args, " ")
	os.Args = append(os.Args, a...)
	c.Assert(config.override(), IsNil)
	c.Assert(config.validate(), IsNil)
	// should be set
	c.Assert(config.Hostname, Equals, "yyy")
	c.Assert(config.PidFile, Equals, "pidfile")
	// --tls should override
	c.Assert(config.Protocol, Equals, "tls")
	c.Assert(config.Poll, Equals, true)
	c.Assert(config.LogLevels, Equals, "<root>=DEBUG")
	// should be overriden
	c.Assert(config.Facility, Equals, syslog.LogUser)
	c.Assert(config.Severity, Equals, syslog.SevNotice)
	c.Assert(config.RefreshInterval, Equals, RefreshInterval(20*time.Second))
	c.Assert(config.Files, HasLen, 3)
	c.Assert(config.Files, DeepEquals, []string{"locallog.txt", "a", "b"})
	c.Assert(config.ConfigFile, Equals, "test/config_with_host.yaml")
	c.Assert(config.DestHost, Equals, "foo")
	c.Assert(config.DestPort, Equals, 1111)
	c.Assert(config.DebugLogFile, Equals, "xxx")
}

func (s *ConfigSuite) TestConfigFileFlag(c *C) {
	config := &Config{
		ConfigFile:      "test/config_with_host.yaml",
		ExcludeFiles:    RegexCollection{},
		ExcludePatterns: RegexCollection{},
		DestHost:        "localhost",
	}
	c.Assert(config.load(), IsNil)
	c.Assert(config.validate(), IsNil)
	// fudge the args
	os.Args = os.Args[0:1]
	args := `-c test/test_config1.yaml`
	a := strings.Split(args, " ")
	os.Args = append(os.Args, a...)
	c.Assert(config.override(), IsNil)
	c.Assert(config.validate(), IsNil)
	c.Assert(config.Hostname, Equals, "baz")
	c.Assert(config.Protocol, Equals, "tcp")
	c.Assert(config.RefreshInterval, Equals, RefreshInterval(20*time.Second))
	c.Assert(config.Files, HasLen, 1)
	c.Assert(config.Files, DeepEquals, []string{"foo.txt"})
	c.Assert(config.ConfigFile, Equals, "test/test_config1.yaml")
	c.Assert(config.DestHost, Equals, "bar")
	c.Assert(config.DestPort, Equals, 515)
}

func (s *ConfigSuite) TestProtocolOverride(c *C) {
	config := &Config{
		ExcludeFiles:    RegexCollection{},
		ExcludePatterns: RegexCollection{},
		Protocol:        "tcp",
	}
	os.Args = os.Args[0:1]
	c.Assert(config.override(), IsNil)
	c.Assert(config.Protocol, Equals, "tcp")
	os.Args = os.Args[0:1]
	os.Args = append(os.Args, `--tls`)
	c.Assert(config.override(), IsNil)
	c.Assert(config.Protocol, Equals, "tls")
	os.Args = os.Args[0:1]
	os.Args = append(os.Args, `--tcp`)
	c.Assert(config.override(), IsNil)
	c.Assert(config.Protocol, Equals, "tcp")

}
