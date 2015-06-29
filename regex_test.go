package main

import (
	. "gopkg.in/check.v1"
	"gopkg.in/yaml.v1"
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
}
