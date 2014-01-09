package main

import (
	"regexp"
	"testing"
)

func TestGetHostnameFromConfig(t *testing.T) {

	cm := ConfigManager{}
	cm.Flags.ConfigFile = "test/config_with_host.yaml"
	cm.loadConfigFile()

	expectedHostname := "test-host-from-config"
	if cm.Hostname() != expectedHostname {
		t.Errorf("Bad hostname, expected %s but got %s", expectedHostname, cm.Hostname())
	}
}

func TestGetHostnameFromCommandline(t *testing.T) {
	cm := ConfigManager{}
	cm.Flags.ConfigFile = "test/config_with_host.yaml"
	cm.loadConfigFile()

	cm.Flags.Hostname = "test-host-from-commandline"

	if cm.Hostname() != cm.Flags.Hostname {
		t.Errorf("Bad hostname, expected %s but got %s", cm.Flags.Hostname, cm.Hostname())
	}
}

func TestFilters(t *testing.T) {
	expressions := []*regexp.Regexp{}
	expressions = append(expressions, regexp.MustCompile("\\d+"))
	message := "test message"
	if matchExps(message, expressions) {
		t.Errorf("Did not expect \"%s\" to match \"%s\"", message, expressions[0])
	}

	message = "0000"
	if !matchExps(message, expressions) {
		t.Errorf("Expected \"%s\" to match \"%s\"", message, expressions[0])
	}
}
