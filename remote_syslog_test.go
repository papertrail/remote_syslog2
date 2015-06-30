package main

import (
	"regexp"

	. "gopkg.in/check.v1"
)

type SyslogSuite struct {
	config *Config
}

var _ = Suite(&SyslogSuite{})

func (s *SyslogSuite) SetUpSuite(c *C) {
	config, err := NewConfig()
	c.Assert(err, IsNil)
	s.config = config
	s.config.ConfigFile = "test/config_with_host.yaml"
	c.Assert(s.config.load(), IsNil)
	c.Assert(s.config.validate(), IsNil)
}

func (s *SyslogSuite) TearDownSuite(c *C) {
}

func (s *SyslogSuite) SetUpTest(c *C) {
}

func (s *SyslogSuite) TearDownTest(c *C) {
}

func (s *SyslogSuite) TestFilters(c *C) {
	expressions := []*regexp.Regexp{regexp.MustCompile("\\d+")}
	message := "test message"
	c.Assert(match(message, expressions), Not(Equals), true)
	message = "0000"
	c.Assert(match(message, expressions), Equals, true)
}
