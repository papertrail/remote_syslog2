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
	"sync"
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

func (c *conn) Close() error {
	return c.netConn.Close()
}

// dial connects to the server and set up a watching goroutine
func dial(network, raddr string, rootCAs *x509.CertPool, connectTimeout time.Duration) (*conn, error) {
	var netConn net.Conn
	var err error

	switch network {
	case "tls":
		var config *tls.Config
		if rootCAs != nil {
			config = &tls.Config{RootCAs: rootCAs}
		}
		dialer := &net.Dialer{
			Timeout:   connectTimeout,
			KeepAlive: time.Second * 60 * 3, // 3 minutes
		}
		netConn, err = tls.DialWithDialer(dialer, "tcp", raddr, config)
	case "udp", "tcp":
		netConn, err = net.DialTimeout(network, raddr, connectTimeout)
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

// A Logger is a connection to a syslog server. It reconnects on error.
// Clients log by sending a Packet to the logger.Packets channel.
type Logger struct {
	conn           *conn
	Packets        chan Packet
	Errors         chan error
	ClientHostname string

	network          string
	raddr            string
	rootCAs          *x509.CertPool
	connectTimeout   time.Duration
	writeTimeout     time.Duration
	tcpMaxLineLength int
	mu               sync.Mutex
	stopChan         chan struct{}
	stopped          bool
}

// Dial connects to the syslog server at raddr, using the optional certBundle,
// and launches a goroutine to watch logger.Packets for messages to log.
func Dial(clientHostname, network, raddr string, rootCAs *x509.CertPool, connectTimeout time.Duration, writeTimeout time.Duration, tcpMaxLineLength int) (*Logger, error) {
	// dial once, just to make sure the network is working
	conn, err := dial(network, raddr, rootCAs, connectTimeout)

	logger := &Logger{
		ClientHostname:   clientHostname,
		network:          network,
		raddr:            raddr,
		rootCAs:          rootCAs,
		Packets:          make(chan Packet, 100),
		Errors:           make(chan error, 0),
		connectTimeout:   connectTimeout,
		writeTimeout:     writeTimeout,
		conn:             conn,
		tcpMaxLineLength: tcpMaxLineLength,
		stopChan:         make(chan struct{}, 1),
	}
	go logger.writeLoop()
	return logger, err
}

func (l *Logger) Write(packet Packet) {
	if l.stopped {
		return
	}

	l.Packets <- packet
}

func (l *Logger) Close() error {
	l.mu.Lock()
	defer l.mu.Unlock()

	if !l.stopped {
		l.stopped = true
		l.stopChan <- struct{}{}

		err := l.conn.Close()
		l.conn = nil

		close(l.Errors)

		return err
	}

	return nil
}

// Connect to the server, retrying every 10 seconds until successful.
func (l *Logger) connect() {
	for {
		c, err := dial(l.network, l.raddr, l.rootCAs, l.connectTimeout)
		if err == nil {
			l.conn = c
			return
		} else {
			l.handleError(err)
			time.Sleep(10 * time.Second)
		}
	}
	panic("unreachable")
}

// Send an error to the Error channel, but don't block if nothing is listening
func (l *Logger) handleError(err error) {
	select {
	case l.Errors <- err:
	default:
	}
}

// Write a packet, reconnecting if needed. It is not safe to call this
// method concurrently.
func (l *Logger) writePacket(p Packet) {
	var err error
	for {
		if l.conn.reconnectNeeded() {
			l.connect()
		}

		deadline := time.Now().Add(l.writeTimeout)
		switch l.conn.netConn.(type) {
		case *net.TCPConn, *tls.Conn:
			l.conn.netConn.SetWriteDeadline(deadline)
			_, err = io.WriteString(l.conn.netConn, p.Generate(l.tcpMaxLineLength)+"\n")
		case *net.UDPConn:
			l.conn.netConn.SetWriteDeadline(deadline)
			_, err = io.WriteString(l.conn.netConn, p.Generate(1024))
		default:
			panic(fmt.Errorf("Network protocol %s not supported", l.network))
		}
		if err == nil {
			return
		} else {
			// We had an error -- we need to close the connection and try again
			l.conn.netConn.Close()
			l.handleError(err)
			time.Sleep(10 * time.Second)
		}
	}
}

// writeloop writes any packets recieved on l.Packets() to the syslog server.
func (l *Logger) writeLoop() {
	for {
		select {
		case p := <-l.Packets:
			l.writePacket(p)
		case <-l.stopChan:
			return
		}
	}
}
