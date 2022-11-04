package main

import (
	"regexp"
	"testing"
	"time"

	"github.com/papertrail/remote_syslog2/syslog"
	"github.com/stretchr/testify/assert"
)

func TestRawConfig(t *testing.T) {
	assert := assert.New(t)
	initConfigAndFlags()

	// pretend like some things were passed on the command line
	flags.Set("configfile", "test/config.yaml")
	flags.Set("tls", "true")

	c, err := NewConfigFromEnv()
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(c.Destination.Host, "logs.papertrailapp.com")
	assert.Equal(c.Destination.Port, 514)
	assert.Equal(c.Destination.Protocol, "tls")
	assert.Equal(c.Destination.Token, "0123456789-ABCDEFGHIJKLMNOPQRSTUVWXYZ_abcdefghijklmnopqrstuvwxyz")
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
	assert.Equal(c.NoDetach, false)
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
}

func TestNoConfigFile(t *testing.T) {
	assert := assert.New(t)
	initConfigAndFlags()

	flags.Set("dest-host", "localhost")
	flags.Set("dest-port", "999")

	c, err := NewConfigFromEnv()
	if err != nil {
		t.Fatal(err)
	}

	assert.NoError(c.Validate())
	assert.Equal("localhost", c.Destination.Host)
	assert.Equal(999, c.Destination.Port)
	assert.Equal("udp", c.Destination.Protocol)
	assert.Equal("", c.Destination.Token)
}

func TestURIInConfig(t *testing.T) {
	assert := assert.New(t)
	initConfigAndFlags()

	flags.Set("dest-uri", "syslog+tls://localhost:999")

	c, err := NewConfigFromEnv()
	if err != nil {
		t.Fatal(err)
	}

	assert.NoError(c.Validate())
	assert.Equal("localhost", c.Destination.Host)
	assert.Equal(999, c.Destination.Port)
	assert.Equal("tls", c.Destination.Protocol)
	assert.Equal("", c.Destination.Token)
}

func TestLogFileTagPatterns(t *testing.T) {
	type tc struct {
		pattern string
		idx     int
		file    string
		tag     string
		err     error
		ok      bool
	}
	tcs := []tc{
		{
			`re:containers/(.*).log=/var/log/containers/*.log`,
			1,
			"/var/log/containers/azure-ip-masq-agent-aaaaa_kube-system_azure-ip-masq-agent-aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa.log",
			"azure-ip-masq-agent-aaaaa_kube-system_azure-ip-masq-agent-aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
			nil,
			true,
		},
		{
			`re:(containers/)(?P<tag>.*)(.log)=/var/log/containers/*.log`,
			2,
			"/var/log/containers/azure-ip-masq-agent-aaaaa_kube-system_azure-ip-masq-agent-aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa.log",
			"azure-ip-masq-agent-aaaaa_kube-system_azure-ip-masq-agent-aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
			nil,
			true,
		},
		{
			`re:containers/(?P<tag>.*)-.{64}\.log=/var/log/containers/*.log`,
			1,
			"/var/log/containers/azure-ip-masq-agent-aaaaa_kube-system_azure-ip-masq-agent-aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa.log",
			"azure-ip-masq-agent-aaaaa_kube-system_azure-ip-masq-agent",
			nil,
			true,
		},
		{
			`re:pods/[^/]+/([^/]+)/=/var/log/pods/*/*/*.log`,
			1,
			"/var/log/pods/kube-system_azure-ip-masq-agent-aaaaa_00000000-0000-0000-0000-000000000000/azure-ip-masq-agent/0.log",
			"azure-ip-masq-agent",
			nil,
			true,
		},
	}

	for _, tc := range tcs {
		t.Run(tc.pattern, func(t *testing.T) {
			t.Helper()

			lfs, err := decodeLogFiles([]interface{}{tc.pattern})
			if tc.err == nil {
				assert.NoError(t, err)
			} else {
				assert.EqualError(t, err, tc.err.Error())
				return
			}
			assert.Len(t, lfs, 1)
			lf := lfs[0]
			assert.NotNil(t, lf.TagPattern)
			assert.Equal(t, lf.TagMatchIndex, tc.idx)
			tag, ok := lfs[0].TagFromFileName(tc.file)
			if tc.ok {
				assert.True(t, ok)
				assert.Equal(t, tag, tc.tag)
			} else {
				assert.False(t, ok)
			}
		})
	}
}
