package main

import (
	"regexp"
	"testing"

	. "gopkg.in/check.v1"
)

func Test(t *testing.T) { TestingT(t) }

type SyslogSuite struct {
	cm *ConfigManager
}

var _ = Suite(&SyslogSuite{})

func (s *SyslogSuite) SetUpSuite(c *C) {
	cm, err := NewConfigManager()
	c.Assert(err, IsNil)
	s.cm = cm
	s.cm.Flags.ConfigFile = "test/config_with_host.yaml"
	c.Assert(s.cm.loadConfigFile(), IsNil)
}

func (s *SyslogSuite) TearDownSuite(c *C) {
}

func (s *SyslogSuite) SetUpTest(c *C) {
}

func (s *SyslogSuite) TearDownTest(c *C) {
}

func (s *SyslogSuite) TestConfig(c *C) {
	c.Assert(s.cm.Hostname(), Equals, "test-host-from-config")
	s.cm.Flags.Hostname = "test-host-from-commandline"
	c.Assert(s.cm.Hostname(), Equals, s.cm.Flags.Hostname)
}

func (s *SyslogSuite) TestFilters(c *C) {
	expressions := []*regexp.Regexp{regexp.MustCompile("\\d+")}
	message := "test message"
	c.Assert(matchExps(message, expressions), Not(Equals), true)
	message = "0000"
	c.Assert(matchExps(message, expressions), Equals, true)
}
