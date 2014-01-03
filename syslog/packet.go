package syslog

import (
	"fmt"
	"strings"
	"time"
)

// A Syslog Priority is a combination of Severity and Facility.
type Priority int

// RFC5424 Severities
const (
	SevEmerg Priority = iota
	SevAlert
	SevCrit
	SevErr
	SevWarning
	SevNotice
	SevInfo
	SevDebug
)

// RFC5424 Facilities
const (
	LogKern Priority = iota
	LogUser
	LogMail
	LogDaemon
	LogAuth
	LogSyslog
	LogLPR
	LogNews
	LogUUCP
	LogCron
	LogAuthPriv
	LogFTP
	LogNTP
	LogAudit
	LogAlert
	LogAt
	LogLocal0
	LogLocal1
	LogLocal2
	LogLocal3
	LogLocal4
	LogLocal5
	LogLocal6
	LogLocal7
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
	ts := p.Time.Format(time.RFC3339Nano)
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
