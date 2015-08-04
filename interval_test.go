package main

import (
	"fmt"
	"time"

	. "gopkg.in/check.v1"
	"gopkg.in/yaml.v2"
)

type IntervalSuite struct {
}

var _ = Suite(&IntervalSuite{})

type i1 struct {
	A RefreshInterval
	B RefreshInterval
	C RefreshInterval
}

func (s *IntervalSuite) SetUpSuite(c *C) {
}

func (s *IntervalSuite) TestInterval(c *C) {
	var data = `
a: 100s
b: 10
c: '20'
`
	v := &i1{}
	c.Assert(yaml.Unmarshal([]byte(data), &v), IsNil)
	c.Assert(v.A, Equals, RefreshInterval(100*time.Second))
	c.Assert(v.B, Equals, RefreshInterval(10*time.Second))
	c.Assert(v.C, Equals, RefreshInterval(20*time.Second))
}

func (s *IntervalSuite) TestMinIntervalInt(c *C) {
	min := time.Duration(MIN_REFRESH_INTERVAL) / time.Second
	data := fmt.Sprintf("a: %d", min-1)
	v := &i1{}
	c.Assert(yaml.Unmarshal([]byte(data), &v), NotNil)
}

func (s *IntervalSuite) TestMinIntervalString(c *C) {
	min := time.Duration(MIN_REFRESH_INTERVAL) / time.Second
	data := fmt.Sprintf("b: '%d'", min-1)
	v := &i1{}
	c.Assert(yaml.Unmarshal([]byte(data), &v), NotNil)
}

func (s *IntervalSuite) TestMinIntervalDuration(c *C) {
	min := time.Duration(MIN_REFRESH_INTERVAL) / time.Second
	data := fmt.Sprintf("c: %ds", min-1)
	v := &i1{}
	c.Assert(yaml.Unmarshal([]byte(data), &v), NotNil)
}
