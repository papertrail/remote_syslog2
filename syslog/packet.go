package syslog

import (
	"fmt"
	"strings"
	"time"
)

// A Packet represents an RFC5425 syslog message
type Packet struct {
	Severity Priority
	Facility Priority
	Hostname string
	Tag      string
	Time     time.Time
	Message  string
}

// like time.RFC3339Nano but with a limit of 6 digits in the SECFRAC part
const rfc5424time = "2006-01-02T15:04:05.999999Z07:00"

// The combined Facility and Severity of this packet. See RFC5424 for details.
func (p Packet) Priority() Priority {
	return (p.Facility << 3) | p.Severity
}

func (p Packet) cleanMessage() string {
	s := strings.Replace(p.Message, "\n", " ", -1)
	s = strings.Replace(s, "\r", " ", -1)
	return strings.Replace(s, "\x00", " ", -1)
}

// Generate creates a RFC5424 syslog format string for this packet.
func (p Packet) Generate(max_size int) string {
	ts := p.Time.Format(rfc5424time)
	if max_size == 0 {
		return fmt.Sprintf("<%d>1 %s %s %s - - - %s", p.Priority(), ts, p.Hostname, p.Tag, p.cleanMessage())
	} else {
		msg := fmt.Sprintf("<%d>1 %s %s %s - - - %s", p.Priority(), ts, p.Hostname, p.Tag, p.cleanMessage())
		if len(msg) > max_size {
			return msg[0:max_size]
		} else {
			return msg
		}
	}
}
