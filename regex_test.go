package main

import (
	. "gopkg.in/check.v1"
	"gopkg.in/yaml.v2"
)

type RegexSuite struct {
}

var _ = Suite(&RegexSuite{})

type r1 struct {
	A RegexCollection
	B RegexCollection
}

func (s *RegexSuite) SetUpSuite(c *C) {
}

func (s *RegexSuite) TestRegex(c *C) {
	var data = `
a: [.*]
b: 
    - .*
    - \s
`
	v := &r1{}
	c.Assert(yaml.Unmarshal([]byte(data), &v), IsNil)
	c.Assert(v.A, HasLen, 1)
	c.Assert(v.B, HasLen, 2)
}
