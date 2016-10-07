package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"regexp"
	"testing"
	"time"

	"github.com/papertrail/remote_syslog2/syslog"
	"github.com/stretchr/testify/assert"
)

const (
	tmpdir     = "./tmp"
	listenHost = "localhost"
	listenPort = 8999
)

var (
	listener *net.UDPConn
)

func init() {
	resAddr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("%s:%d", listenHost, listenPort))
	if err != nil {
		panic(err)
	}

	listener, err = net.ListenUDP("udp", resAddr)
	if err != nil {
		panic(err)
	}
}

// main testing function to clean up after running
func TestMain(m *testing.M) {
	os.Mkdir(tmpdir, 0755)
	rs := m.Run()
	os.RemoveAll(tmpdir)
	os.Exit(rs)
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

func TestNewFileSeek(t *testing.T) {
	assert := assert.New(t)

	s := NewServer(testConfig())
	go s.Start()
	defer s.Close()

	// just a quick rest to get the server started
	time.Sleep(1 * time.Second)

	for _, msg := range []string{
		"welcome to the jungle",
		"we got alerts and logs",
		"we got everything you want",
		"as long as it's alerts and logs",
	} {
		file := tmpLogFile()
		defer file.Close()

		writeLog(file, msg)

		packet := readPacket()
		assert.Equal(msg, packet.Message)
	}
}

// write to test log file
func writeLog(file *os.File, msg string) {
	w := bufio.NewWriterSize(file, 1024*32)

	if _, err := w.WriteString(msg + "\n"); err != nil {
		panic(err)
	}

	w.Flush()
}

// creates a log file that matches our pattern (tmp/*.log)
func tmpLogFile() *os.File {
	file, err := os.Create(fmt.Sprintf("tmp/%d.log", time.Now().UnixNano()))
	if err != nil {
		panic(err)
	}

	return file
}

func readPacket() syslog.Packet {
	listener.SetReadDeadline(time.Now().Add(1200 * time.Millisecond))
	reader := bufio.NewReaderSize(listener, 1024*32)

	line, prefix, err := reader.ReadLine()
	if prefix {
		panic("reader buffer too small")
	}
	if err != nil {
		panic(err)
	}

	packet, err := syslog.Parse(string(line))
	if err != nil {
		panic(err)
	}

	return packet
}

func testConfig() *Config {
	severity, _ := syslog.Severity("info")
	facility, _ := syslog.Facility("user")

	return &Config{
		ConnectTimeout:       10 * time.Second,
		WriteTimeout:         10 * time.Second,
		NewFileCheckInterval: 1 * time.Second,
		LogLevels:            "<root>=INFO",
		TcpMaxLineLength:     99990,
		NoDetach:             true,
		Hostname:             "testhost",
		Severity:             severity,
		Facility:             facility,
		Destination: struct {
			Host     string
			Port     int
			Protocol string
		}{
			Host:     listenHost,
			Port:     listenPort,
			Protocol: "udp",
		},
		Files: []LogFile{
			{
				Path: "tmp/*.log",
			},
		},
	}
}
