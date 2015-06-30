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
	c.Assert(config.validate(), IsNil)
	// fudge the args
	os.Args = os.Args[0:1]
	args := `-c test/stuff.yaml -d foo -p 1111 --debug-log-cfg=xxx --new-file-check-interval=20s --hostname=yyy --pid-file=pidfile --tcp --tls --poll --log=<root>=DEBUG a b`
	a := strings.Split(args, " ")
	os.Args = append(os.Args, a...)
	c.Assert(config.override(), IsNil)
	c.Assert(config.validate(), IsNil)
	// should be set
	c.Assert(config.Hostname, Equals, "yyy")
	c.Assert(config.PidFile, Equals, "pidfile")
	c.Assert(config.UseTCP, Equals, true)
	c.Assert(config.UseTLS, Equals, true)
	c.Assert(config.Poll, Equals, true)
	c.Assert(config.LogLevels, Equals, "<root>=DEBUG")
	// should be overriden
	c.Assert(config.Facility, Equals, syslog.LogUser)
	c.Assert(config.Severity, Equals, syslog.SevNotice)
	c.Assert(config.RefreshInterval, Equals, RefreshInterval(20*time.Second))
	c.Assert(config.Files, HasLen, 3)
	c.Assert(config.Files, DeepEquals, []string{"locallog.txt", "a", "b"})
	c.Assert(config.ConfigFile, Equals, "test/stuff.yaml")
	c.Assert(config.DestHost, Equals, "foo")
	c.Assert(config.DestPort, Equals, 1111)
	c.Assert(config.DebugLogFile, Equals, "xxx")
}
