package syslog

import (
	"crypto/tls"
	"fmt"
	"github.com/sevenscale/remote_syslog2/syslog/certs"
	"io"
	"net"
	"time"
)

// A net.Conn with added reconnection logic
type conn struct {
	netConn net.Conn
	errors  chan error
}

// watch watches the connection for error, sends detected error to c.errors
func (c *conn) watch() {
	for {
		data := make([]byte, 1)
		_, err := c.netConn.Read(data)
		if err != nil {
			c.netConn.Close()
			c.errors <- err
			return
		}
	}
}

// reconnectNeeded determines if a reconnect is needed by checking for a
// message on the readErrors channel
func (c *conn) reconnectNeeded() bool {
	if c == nil {
		return true
	}
	select {
	case <-c.errors:
		return true
	default:
		return false
	}
}

// dial connects to the server and set up a watching goroutine
func dial(network, raddr string, bundle *certs.CertBundle) (*conn, error) {
	var netConn net.Conn
	var err error

	switch network {
	case "tls":
		config := tls.Config{InsecureSkipVerify: false, RootCAs: &(*bundle).CertPool}
		netConn, err = tls.Dial("tcp", raddr, &config)
	case "udp", "tcp":
		netConn, err = net.Dial(network, raddr)
	default:
		return nil, fmt.Errorf("Network protocol %s not supported", network)
	}
	if err != nil {
		return nil, err
	} else {
		c := &conn{netConn, make(chan error)}
		go c.watch()
		return c, nil
	}
}

// Connect to the server, retrying until successful
func connect(network, raddr string, bundle *certs.CertBundle) *conn {
	for {
		c, err := dial(network, raddr, bundle)
		if err == nil {
			return c
		} else {
			time.Sleep(5 * time.Second)
		}
	}
	panic("unreachable")
}

// A Logger is a connection to a syslog server. It reconnects on error.
// Clients log by sending a Packet to the logger.Packets channel.
type Logger struct {
	conn           *conn
	Packets        chan Packet
	ClientHostname string

	network    string
	raddr      string
	certBundle *certs.CertBundle
}

// Dial connects to the syslog server at raddr, using the optional certBundle,
// and launches a goroutine to watch logger.Packets for messages to log.
func Dial(clientHostname, network, raddr string, certBundle *certs.CertBundle) (*Logger, error) {
	// dial once, just to make sure the network is working
	conn, err := dial(network, raddr, certBundle)

	if err != nil {
		return nil, err
	} else {
		logger := &Logger{
			ClientHostname: clientHostname,
			network:        network,
			raddr:          raddr,
			certBundle:     certBundle,
			Packets:        make(chan Packet, 100),
			conn:           conn,
		}
		go logger.writeLoop()
		return logger, nil
	}
}

// Write a packet, reconnecting if needed. It is not safe to call this
// method concurrently.
func (l *Logger) writePacket(p Packet) (err error) {
	if l.conn.reconnectNeeded() {
		l.conn = connect(l.network, l.raddr, l.certBundle)
	}

	switch l.conn.netConn.(type) {
	case *net.TCPConn, *tls.Conn:
		_, err = io.WriteString(l.conn.netConn, p.Generate(0)+"\n")
		return err
	case *net.UDPConn:
		_, err = io.WriteString(l.conn.netConn, p.Generate(1024))
		return err
	default:
		panic(fmt.Errorf("Network protocol %s not supported", l.network))
	}
}

// writeloop writes any packets recieved on l.Packets() to the syslog server.
func (l *Logger) writeLoop() {
	for p := range l.Packets {
		l.writePacket(p)
	}
}
