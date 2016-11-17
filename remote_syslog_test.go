package main

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/papertrail/remote_syslog2/syslog"
	"github.com/stretchr/testify/assert"
)

const (
	tmpdir = "./tmp"
)

var (
	server *testSyslogServer
)

// main testing function to clean up after running
func TestMain(m *testing.M) {
	os.Mkdir(tmpdir, 0755)

	server = newTestSyslogServer("127.0.0.1:0")
	go server.serve()

	rs := m.Run()

	server.closeCh <- struct{}{}

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

		packet := <-server.packets
		assert.Equal(msg, packet.Message)
	}
}

func TestGlobCollisions(t *testing.T) {
	assert := assert.New(t)

	// Make sure we're running on a clean directory
	os.RemoveAll(tmpdir)
	os.Mkdir(tmpdir, 0755)

	// Add colliding globs
	config := testConfig()
	config.Files = append(config.Files, LogFile{
		Path: "tmp/*.log",
	})

	// Use an observable registry
	testRegistry := &testRegistry{workers: make(map[string]int)}

	s := NewServer(config)
	s.registry = testRegistry
	go s.Start()
	defer s.Close()

	// just a quick rest to get the server started
	time.Sleep(1 * time.Second)

	var files []*os.File
	for i := 0; i < 50; i++ {
		file := tmpLogFile()
		files = append(files, file)
		writeLog(file, "the most important message"+strconv.Itoa(i))
	}

	// NewFileCheckInterval = 1 second, so wait 1100ms for messages
	time.Sleep(3000 * time.Millisecond)

	testRegistry.mu.RLock()
	for file, forwardCount := range testRegistry.workers {
		assert.Equal(1, forwardCount, "Expected %s to be added once, got %d", file, forwardCount)
	}
	testRegistry.mu.RUnlock()

	for _, file := range files {
		file.Close()
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

func testConfig() *Config {
	severity, _ := syslog.Severity("info")
	facility, _ := syslog.Facility("user")
	addr := server.addr()

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
			Host:     addr.host,
			Port:     addr.port,
			Protocol: "tcp",
		},
		Files: []LogFile{
			{
				Path: "tmp/*.log",
			},
		},
	}
}

type testSyslogServer struct {
	listener net.Listener
	closeCh  chan struct{}
	packets  chan syslog.Packet
}

type testAddr struct {
	host string
	port int
}

// testSyslogServer is a type for testing sent syslog messages
func newTestSyslogServer(addr string) *testSyslogServer {
	var (
		err error
		s   = &testSyslogServer{
			closeCh: make(chan struct{}),
			packets: make(chan syslog.Packet),
		}
	)

	s.listener, err = net.Listen("tcp", addr)
	if err != nil {
		panic(err)
	}

	return s
}

func (s *testSyslogServer) addr() testAddr {
	addr := s.listener.Addr()
	addrParts := strings.Split(addr.String(), ":")
	if len(addrParts) != 2 {
		panic(fmt.Sprintf("invalid listener address: %s", addr.String()))
	}

	port, err := strconv.Atoi(addrParts[1])
	if err != nil {
		panic(err)
	}

	return testAddr{
		host: addrParts[0],
		port: port,
	}
}

func (s *testSyslogServer) serve() {
	for {
		select {
		case <-s.closeCh:
			return

		default:
			conn, err := s.listener.Accept()
			if err != nil {
				panic(err)
			}

			reader := bufio.NewReader(conn)
			for {
				select {
				case <-s.closeCh:
					return

				default:
					line, err := reader.ReadString('\n')
					if err != nil && err != io.EOF {
						panic(err)
					}

					if err == io.EOF {
						time.Sleep(100 * time.Millisecond)
						continue
					}

					fmt.Printf(line)
					packet, err := syslog.Parse(strings.TrimRight(line, "\n"))
					if err != nil {
						panic(err)
					}

					select {
					case s.packets <- packet:
					case <-time.After(time.Second):
					}
				}
			}
		}
	}
}

// testRegistry is a WorkerRegistry implementation that keeps track of how many times a file was added
type testRegistry struct {
	mu      sync.RWMutex
	workers map[string]int
}

func (tr *testRegistry) Exists(worker string) bool {
	tr.mu.RLock()
	defer tr.mu.RUnlock()
	_, ok := tr.workers[worker]
	return ok
}

func (tr *testRegistry) Add(worker string) {
	tr.mu.Lock()
	defer tr.mu.Unlock()
	log.Tracef("Adding %s to worker registry", worker)
	tr.workers[worker] += 1
}

func (tr *testRegistry) Remove(worker string) {
	tr.mu.Lock()
	defer tr.mu.Unlock()
	log.Tracef("Removing %s from worker registry", worker)
	delete(tr.workers, worker)
}
