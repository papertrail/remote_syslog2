package main

import (
	"fmt"
	"regexp"
)

type RegexCollection []*regexp.Regexp

func (self *RegexCollection) String() string {
	return fmt.Sprint(*self)
}

func (self *RegexCollection) UnmarshalYAML(
	unmarshal func(interface{}) error,
) error {
	a := []string{}
	if err := unmarshal(&a); err != nil {
		return fmt.Errorf("Expected string array: %v", err)
	}
	for _, v := range a {
		exp, err := regexp.Compile(v)
		if err != nil {
			return err
		}
		*self = append(*self, exp)
	}
	return nil
}
