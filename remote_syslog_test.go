package main

import (
	"encoding/json"
	"io/ioutil"
	"testing"
)

func TestGetHostnameFromConfig(t *testing.T) {
	cf, err := loadConfigFile("test/config_with_host.json")
	if err != nil {
		t.Errorf("Config error %s", err)
	}
	expectedHostname := "test-host-from-config"
	if cf.GetHostname() != expectedHostname {
		t.Errorf("Bad hostname, expected %s but got %s", expectedHostname, cf.GetHostname())
	}
}

func TestGetHostnameFromCommandline(t *testing.T) {
	cf, err := loadConfigFile("test/config_with_host.json")
	if err != nil {
		t.Errorf("Config error %s", err)
	}
	//Set the commandline flag to the desired hostname
	configHostname = "test-host-from-commandline"
	if cf.GetHostname() != configHostname {
		t.Errorf("Bad hostname, expected %s but got %s", configHostname, cf.GetHostname())
	}
}

func loadConfigFile(configFile string) (ConfigFile, error) {
	var config ConfigFile
	file, err := ioutil.ReadFile(configFile)
	if err != nil {
		return config, err
	}

	err = json.Unmarshal(file, &config)
	if err != nil {
		return config, err
	}

	return config, nil
}
