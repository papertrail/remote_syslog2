package main

import (
	"io/ioutil"
	"os"
	"path"
	"regexp"
	"time"

	"github.com/papertrail/remote_syslog2/syslog"
	. "gopkg.in/check.v1"
)

type TailSuite struct {
	wr      *WorkerRegistry
	tempdir string
	logger  *MockLogger
}

var _ = Suite(&TailSuite{})

func (s *TailSuite) SetUpSuite(c *C) {
	s.wr = NewWorkerRegistry()
	s.logger = NewMockLogger()
}

func (s *TailSuite) TearDownSuite(c *C) {
}

func (s *TailSuite) SetUpTest(c *C) {
}

func (s *TailSuite) TearDownTest(c *C) {
	os.RemoveAll(s.tempdir)
}

func (s *TailSuite) TestSimple(c *C) {
	dir, err := ioutil.TempDir("", "TailSimple")
	c.Assert(err, IsNil)
	s.tempdir = dir
	f, err := ioutil.TempFile(s.tempdir, "")
	c.Assert(err, IsNil)
	data := `one
two
three
four
`
	c.Assert(ioutil.WriteFile(f.Name(), []byte(`start`), os.ModePerm), IsNil)
	pats := []*regexp.Regexp{regexp.MustCompile("three")}
	go tailone(
		f.Name(),
		pats,
		syslog.SevNotice,
		syslog.LogLocal4,
		false,
		s.logger,
		s.wr,
	)
	time.Sleep(1 * time.Second)
	s.append(f.Name(), data, c)
	p := s.packet(f, c)
	c.Assert(p.Message, Equals, `one`)
	p = s.packet(f, c)
	c.Assert(p.Message, Equals, `two`)
	p = s.packet(f, c)
	c.Assert(p.Message, Equals, `four`)
}

func (s *TailSuite) packet(f *os.File, c *C) *syslog.Packet {
	var p *syslog.Packet
	select {
	case p = <-s.logger.Packets():
	case <-time.After(1 * time.Second):
		c.Fatal("Out of packets")
	}
	c.Assert(p.Severity, Equals, syslog.SevNotice)
	c.Assert(p.Facility, Equals, syslog.LogLocal4)
	c.Assert(p.Hostname, Equals, "mocklogger")
	c.Assert(p.Tag, Equals, path.Base(f.Name()))
	return p
}

func (s *TailSuite) append(name string, data string, c *C) {
	f, err := os.OpenFile(name, os.O_APPEND|os.O_WRONLY, 0600)
	c.Assert(err, IsNil)
	defer f.Close()
	_, err = f.WriteString(data)
	c.Assert(err, IsNil)
}

// Logger for use in testing
type MockLogger struct {
	packets chan *syslog.Packet
	errors  chan error
}

func NewMockLogger() *MockLogger {
	return &MockLogger{
		packets: make(chan *syslog.Packet, 16),
		errors:  make(chan error, 16),
	}
}

func (self *MockLogger) Packets() chan *syslog.Packet {
	return self.packets
}

func (self *MockLogger) Errors() chan error {
	return self.errors
}

func (self *MockLogger) Hostname() string {
	return "mocklogger"
}
