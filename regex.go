package main

import (
	"fmt"
	"regexp"
)

type RegexCollection []*regexp.Regexp

func (self *RegexCollection) Set(v string) error {
	exp, err := regexp.Compile(v)
	if err != nil {
		return err
	}
	*self = append(*self, exp)
	return nil
}

func (self *RegexCollection) String() string {
	return fmt.Sprint(*self)
}

func (self *RegexCollection) SetYAML(tag string, v interface{}) bool {
	a, ok := v.([]interface{})
	if !ok {
		return false
	}
	for _, item := range a {
		s, ok := item.(string)
		if !ok {
			return false
		}
		if err := self.Set(s); err != nil {
			return false
		}
	}
	return true
}
