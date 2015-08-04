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

func (self *RefreshInterval) UnmarshalYAML(
	unmarshal func(interface{}) error,
) (e error) {
	defer func() {
		if *self < MIN_REFRESH_INTERVAL {
			e = fmt.Errorf("Refresh interval must be greater than or equal to %s", MIN_REFRESH_INTERVAL)
		}
	}()
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
		*self = RefreshInterval(d)
		return nil
	}
	return fmt.Errorf("Invalid refresh interval: %s", s)
}
