package syslog

import (
	"fmt"
	"math/rand"
	"net"
	"testing"
	"time"
)

const clienthost = "clienthost"

func panicf(s string, i ...interface{}) { panic(fmt.Sprintf(s, i)) }

type testServer struct {
	ln       net.Listener
	Addr     string
	Close    chan bool
	Messages chan string
}

func newTestServer() *testServer {
	server := testServer{
		Close:    make(chan bool, 1),
		Messages: make(chan string, 20),
	}
	server.listen()
	go server.serve()
	return &server
}

func (s *testServer) listen() {
	var err error
	if s.Addr == "" {
		s.ln, err = net.Listen("tcp", "127.0.0.1:0")
	} else {
		s.ln, err = net.Listen("tcp", s.Addr)
	}
	if err != nil {
		panicf("listen error %v", err)
	}
	if s.Addr == "" {
		s.Addr = s.ln.Addr().String()
	}
}

func (s testServer) handle(conn net.Conn) {
	for {
		buf := make([]byte, 1024)
		n, err := conn.Read(buf)
		if err != nil {
			panicf("Read error")
		} else {
			s.Messages <- string(buf[0:n])
		}
		// todo: make configurable
		if 0 == (rand.Int() % 2) {
			conn.Close()
			return
		}
	}
}

func (s testServer) serve() {
	for {
		select {
		case <-s.Close:
			s.ln.Close()
			return
		default:
			conn, err := s.ln.Accept()
			if err != nil {
				panicf("Accept error: %v", err)
			}
			go s.handle(conn)
		}
	}
}

func generatePackets() []Packet {
	packets := make([]Packet, 10)
	for i, _ := range packets {
		t, _ := time.Parse(time.RFC3339, "2006-01-02T15:04:05Z07:00")
		packets[i] = Packet{
			Severity: SevInfo,
			Facility: LogLocal1,
			Time:     t,
			Hostname: clienthost,
			Tag:      "test",
			Message:  fmt.Sprintf("message %d", i),
		}
	}
	return packets
}

func TestSyslog(t *testing.T) {
	s := newTestServer()

	logger, err := Dial(clienthost, "tcp", s.Addr, nil)
	if err != nil {
		t.Errorf("unexpected dial error %v", err)
	}
	packets := generatePackets()
	for _, p := range packets {
		logger.writePacket(p)
		time.Sleep(100 * time.Millisecond)
	}
	s.Close <- true

	for _, p := range packets {
		expected := p.Generate(0) + "\n"
		select {
		case got := <-s.Messages:
			if got != expected {
				t.Errorf("expected %s, got %s", expected, got)
			}
		default:
			t.Errorf("expected %s, got nothing", expected)
			break
		}
	}
	if l := len(s.Messages); l != 0 {
		t.Errorf("found %d extra messages", l)
	}
}
