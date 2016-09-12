package main

import (
	"regexp"
	"testing"
	"time"

	"github.com/papertrail/remote_syslog2/papertrail"
	"github.com/papertrail/remote_syslog2/syslog"
	"github.com/spf13/pflag"
	"github.com/stretchr/testify/assert"
)

func TestRawConfig(t *testing.T) {
	assert := assert.New(t)

	// pretend like some things were passed on the command line
	pflag.Set("configfile", "example_config.yml")
	pflag.Set("tls", "true")

	c, err := NewConfigFromEnv()
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(c.Destination.Host, "logs.papertrailapp.com")
	assert.Equal(c.Destination.Port, 514)
	assert.Equal(c.Destination.Protocol, "tls")
	assert.Equal(c.ExcludePatterns, []*regexp.Regexp{regexp.MustCompile("don't log on me"), regexp.MustCompile(`do \w+ on me`)})
	assert.Equal(c.ExcludeFiles, []*regexp.Regexp{regexp.MustCompile(`\.DS_Store`)})
	assert.Equal(c.Files, []LogFile{
		{
			Path: "locallog.txt",
		},
		{
			Path: "/var/log/**/*.log",
		},
		{
			Tag:  "nginx",
			Path: "/var/log/nginx/nginx.log",
		},
		{
			Tag:  "apache",
			Path: "/var/log/httpd/access_log",
		},
	})
	assert.Equal(c.TcpMaxLineLength, 99991)
	assert.Equal(c.NewFileCheckInterval, 10*time.Second)
	assert.Equal(c.ConnectTimeout, 5*time.Second)
	assert.Equal(c.WriteTimeout, 30*time.Second)
	assert.Equal(c.TCP, false)
	assert.Equal(c.TLS, true)
	assert.Equal(c.LogLevels, "<root>=INFO")
	assert.Equal(c.PidFile, "/var/run/rs2.pid")
	assert.Equal(c.DebugLogFile, "/dev/null")
	assert.Equal(c.NoDaemonize, false)
	sev, err := syslog.Severity("notice")
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(c.Severity, sev)
	fac, err := syslog.Facility("user")
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(c.Facility, fac)
	assert.NotEqual(c.Hostname, "")
	assert.Equal(c.Poll, false)
	assert.Equal(c.RootCAs, papertrail.RootCA())
}
