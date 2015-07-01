package main

import (
	"fmt"
	"strconv"
	"time"
)

type RefreshInterval time.Duration

func (self RefreshInterval) String() string {
	return time.Duration(self).String()
}

type Unmarshaler interface {
	UnmarshalYAML(unmarshal func(interface{}) error) error
}

func (self *RefreshInterval) UnmarshalYAML(
	unmarshal func(interface{}) error,
) error {
	i := 0
	if err := unmarshal(&i); err == nil {
		*self = RefreshInterval(time.Duration(i) * time.Second)
		return nil
	}
	s := ""
	if err := unmarshal(&s); err == nil {
		i, err := strconv.ParseUint(s, 10, 64)
		if err == nil {
			*self = RefreshInterval(time.Duration(i) * time.Second)
			return nil
		}
		d, err := time.ParseDuration(s)
		if err != nil {
			return err
		}
		ii := RefreshInterval(d)
		if ii < MIN_REFRESH_INTERVAL {
			return fmt.Errorf("Refresh interval must be greater than or equal to %s", MIN_REFRESH_INTERVAL)
		}
		*self = ii
		return nil
	}
	return fmt.Errorf("Invalid refresh interval: %s", s)
}
