package syslog

import (
	"fmt"
)

// A Syslog Priority is a combination of Severity and Facility.
type Priority int

// Returned when looking up a non-existant facility or severity
var ErrPriority = fmt.Errorf("Not a designated priority")

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

var severities = map[string]Priority{
	"emerg":  SevEmerg,
	"alert":  SevAlert,
	"crit":   SevCrit,
	"err":    SevErr,
	"warn":   SevWarning,
	"notice": SevNotice,
	"info":   SevInfo,
	"debug":  SevDebug,
}

// Severity returns the named severity. It returns ErrPriority if the severity
// does not exist.
func Severity(name string) (Priority, error) {
	p, ok := severities[name]
	if !ok {
		return 0, ErrPriority
	}
	return p, nil
}

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

var facilities = map[string]Priority{
	"kern":     LogKern,
	"user":     LogUser,
	"mail":     LogMail,
	"daemon":   LogDaemon,
	"auth":     LogAuth,
	"syslog":   LogSyslog,
	"lpr":      LogLPR,
	"news":     LogNews,
	"uucp":     LogUUCP,
	"cron":     LogCron,
	"authpriv": LogAuthPriv,
	"ftp":      LogFTP,
	"ntp":      LogNTP,
	"audit":    LogAudit,
	"alert":    LogAlert,
	"at":       LogAt,
	"local0":   LogLocal0,
	"local1":   LogLocal1,
	"local2":   LogLocal2,
	"local3":   LogLocal3,
	"local4":   LogLocal4,
	"local5":   LogLocal5,
	"local6":   LogLocal6,
	"local7":   LogLocal7,
}

// Facility returns the named facility. It returns ErrPriority if the facility
// does not exist.
func Facility(name string) (Priority, error) {
	p, ok := facilities[name]
	if !ok {
		return 0, ErrPriority
	}
	return p, nil
}
