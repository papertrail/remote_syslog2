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

func (self *RefreshInterval) Set(v interface{}) error {
	switch v := v.(type) {
	case int:
		*self = RefreshInterval(time.Duration(v) * time.Second)
		return nil
	case string:
		i, err := strconv.ParseUint(v, 10, 64)
		if err == nil {
			*self = RefreshInterval(time.Duration(i) * time.Second)
			return nil
		}
		d, err := time.ParseDuration(v)
		if err != nil {
			return err
		}
		ii := RefreshInterval(d)
		if ii < MinimumRefreshInterval {
			return fmt.Errorf("Refresh interval must be greater than or equal to %s", MinimumRefreshInterval)
		}
		*self = ii
		return nil
	default:
		return fmt.Errorf("Invalid refresh interval: %v", v)
	}
}

func (self *RefreshInterval) SetYAML(tag string, v interface{}) bool {
	if err := self.Set(v); err != nil {
		return false
	}
	return true
}
