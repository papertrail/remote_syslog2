/*
The syslog package provides a syslog client.

Unlike the core log/syslog package it uses the newer rfc5424 syslog protocol,
reliably reconnects on failure, and supports TLS encrypted TCP connections.
*/
package syslog

import (
	_ "crypto/sha512"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io"
	"net"
	"time"
)

type Logger interface {
	Packets() chan *Packet
	Errors() chan error
	Hostname() string
}

// A Logger is a connection to a syslog server. It reconnects on error.
// Clients log by sending a Packet to the logger.Packets channel.
type RemoteLogger struct {
	conn     *conn
	packets  chan *Packet
	errors   chan error
	hostname string
	network  string
	raddr    string
	rootCAs  *x509.CertPool
}

// Dial connects to the syslog server at raddr, using the optional certBundle,
// and launches a goroutine to watch logger.Packets for messages to log.
func Dial(
	clientHostname, network, raddr string,
	rootCAs *x509.CertPool,
) (*RemoteLogger, error) {
	// dial once, just to make sure the network is working
	conn, err := dial(network, raddr, rootCAs)
	if err != nil {
		return nil, err
	}
	logger := &RemoteLogger{
		hostname: clientHostname,
		network:  network,
		raddr:    raddr,
		rootCAs:  rootCAs,
		packets:  make(chan *Packet, 100),
		errors:   make(chan error, 0),
		conn:     conn,
	}
	go logger.writeLoop()
	return logger, nil
}

func (l *RemoteLogger) Packets() chan *Packet {
	return l.packets
}

func (l *RemoteLogger) Errors() chan error {
	return l.errors
}

func (l *RemoteLogger) Hostname() string {
	return l.hostname
}

// Connect to the server, retrying every 10 seconds until successful.
func (l *RemoteLogger) connect() {
	for {
		c, err := dial(l.network, l.raddr, l.rootCAs)
		if err == nil {
			l.conn = c
			return
		}
		l.handleError(err)
		time.Sleep(10 * time.Second)
	}
	panic("unreachable")
}

// Send an error to the Error channel, but don't block if nothing is listening
func (l *RemoteLogger) handleError(err error) {
	select {
	case l.errors <- err:
	default:
	}
}

// Write a packet, reconnecting if needed. It is not safe to call this
// method concurrently.
func (l *RemoteLogger) writePacket(p *Packet) {
	var err error
	for {
		if l.conn.reconnectNeeded() {
			l.connect()
		}
		switch l.conn.netConn.(type) {
		case *net.TCPConn, *tls.Conn:
			_, err = io.WriteString(l.conn.netConn, p.Generate(0)+"\n")
		case *net.UDPConn:
			_, err = io.WriteString(l.conn.netConn, p.Generate(1024))
		default:
			panic(fmt.Errorf("Network protocol %s not supported", l.network))
		}
		if err == nil {
			return
		}
		l.handleError(err)
		time.Sleep(10 * time.Second)
	}
}

// writeloop writes any packets recieved on l.Packets() to the syslog server.
func (l *RemoteLogger) writeLoop() {
	for p := range l.packets {
		l.writePacket(p)
	}
}

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
func dial(network, raddr string, rootCAs *x509.CertPool) (*conn, error) {
	var netConn net.Conn
	var err error

	switch network {
	case "tls":
		var config *tls.Config
		if rootCAs != nil {
			config = &tls.Config{RootCAs: rootCAs}
		}
		netConn, err = tls.Dial("tcp", raddr, config)
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
