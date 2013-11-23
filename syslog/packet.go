// This package contains an implementation of the syslog protocol.
//
// Unlike the core log/syslog package it uses the newer rfc5424 syslog
// protocol.
package syslog

import (
	"fmt"
	"strings"
	"time"
)

// A Syslog Priority is a combination of Severity and Facility.
type Priority int

// Severities
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

// Facilities
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

type Packet struct {
	Severity Priority
	Facility Priority
	Hostname string
	Tag      string
	Time     time.Time
	Message  string
}

func (p Packet) Priority() Priority {
	return (p.Facility << 3) | p.Severity
}

func (p Packet) CleanMessage() string {
	s := strings.Replace(p.Message, "\n", " ", -1)
	return strings.Replace(s, "\x00", " ", -1)
}

func (p Packet) Generate(max_size int) string {
	ts := p.Time.Format(time.RFC3339Nano)
	if max_size == 0 {
		return fmt.Sprintf("<%d>1 %s %s %s - - - %s", p.Priority(), ts, p.Hostname, p.Tag, p.CleanMessage())
	} else {
		msg := fmt.Sprintf("<%d>1 %s %s %s - - - %s", p.Priority(), ts, p.Hostname, p.Tag, p.CleanMessage())
		if len(msg) > max_size {
			return msg[0:max_size]
		} else {
			return msg
		}
	}
}
