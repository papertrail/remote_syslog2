package main

import (
	"regexp"

	. "gopkg.in/check.v1"
)

type SyslogSuite struct {
}

var _ = Suite(&SyslogSuite{})

func (s *SyslogSuite) SetUpSuite(c *C) {
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
