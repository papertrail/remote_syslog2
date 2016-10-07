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

// A convenience function for testing
func Parse(line string) (Packet, error) {
	var (
		packet   Packet
		priority int
		ts       string
		hostname string
		tag      string
	)

	splitLine := strings.Split(line, " - - - ")
	if len(splitLine) != 2 {
		return packet, fmt.Errorf("couldn't parse %s", line)
	}

	fmt.Sscanf(splitLine[0], "<%d>1 %s %s %s", &priority, &ts, &hostname, &tag)

	t, err := time.Parse(rfc5424time, ts)
	if err != nil {
		return packet, err
	}

	return Packet{
		Severity: Priority(priority & 7),
		Facility: Priority(priority >> 3),
		Hostname: hostname,
		Tag:      tag,
		Time:     t,
		Message:  splitLine[1],
	}, nil
}
