package main

import (
	"testing"

	"github.com/spf13/pflag"
	"github.com/stretchr/testify/assert"
)

func TestRawConfig(t *testing.T) {
	assert := assert.New(t)

	// pretend like some things were passed on the command line
	pflag.Set("configfile", "example_config.yml")
	pflag.Set("tls", "true")

	c, err := NewConfig()
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(c.Destination.Host, "logs.papertrailapp.com")
	assert.Equal(c.Destination.Port, 514)
	assert.Equal(c.Destination.Protocol, "tls")
	assert.Equal(c.ExcludePatterns, []string{"don't log on me", `do \w+ on me`})
	assert.Equal(c.ExcludeFiles, []string{`\.DS_Store`})
	assert.Equal(c.Files, []interface{}{
		"locallog.txt",
		"/var/log/**/*.log",
		"nginx=/var/log/nginx/nginx.log",
		map[interface{}]interface{}{"path": "/var/log/httpd/access_log", "tag": "apache"},
	})
	assert.Equal(c.TcpMaxLineLength, 99991)
	assert.Equal(c.NewFileCheckInterval, 10)
	assert.Equal(c.ConnectTimeout, 5)
	assert.Equal(c.WriteTimeout, 0)
	assert.Equal(c.TCP, false)
	assert.Equal(c.TLS, true)
	assert.Equal(c.LogLevels, "<root>=INFO")
	assert.Equal(c.PidFile, "/var/run/rs2.pid")
	assert.Equal(c.DebugLogFile, "/dev/null")
	assert.Equal(c.NoDaemonize, false)
	assert.Equal(c.Severity, "notice")
	assert.Equal(c.Facility, "user")
	assert.NotEqual(c.Hostname, "")
	assert.Equal(c.Poll, false)
}

func TestComputedConfig(t *testing.T) {
	assert := assert.New(t)

	// pretend like some things were passed on the command line
	pflag.Set("configfile", "example_config.yml")

	c, err := NewConfig()
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal("tls", c.GetDestinationProtocol())
	assert.Equal([]LogFile{
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
	}, c.GetFiles())
}
