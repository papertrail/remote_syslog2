package syslog

import (
	"crypto/tls"
	"fmt"
	"github.com/sevenscale/remote_syslog2/syslog/certs"
	"io"
	"net"
)

type Conn struct {
	hostname string
	conn     net.Conn
}

func Dial(network, raddr, hostname string, bundle *certs.CertBundle) (*Conn, error) {
	var conn net.Conn
	var err error

	switch network {
	case "tls":
		config := tls.Config{InsecureSkipVerify: false, RootCAs: &(*bundle).CertPool}
		conn, err = tls.Dial("tcp", raddr, &config)
	case "udp", "tcp":
		conn, err = net.Dial(network, raddr)
	default:
		err = fmt.Errorf("Network protocol %s not supported", network)
	}
	if err != nil {
		return nil, err
	} else {
		if hostname == "" {
			hostname, _, err = net.SplitHostPort(conn.LocalAddr().String())
			if err != nil {
				return nil, err
			}
		}
		return &Conn{conn: conn, hostname: hostname}, nil
	}
}

func (c Conn) Hostname() string {
	return c.hostname
}

func (c Conn) WritePacket(p Packet) (err error) {
	// todo: max size?
	line := p.Generate()

	switch c.conn.(type) {
	case *net.TCPConn, *tls.Conn:
		_, err = io.WriteString(c.conn, line+"\n")
		return err
	case *net.UDPConn:
		_, err = io.WriteString(c.conn, line)
		return err
	default:
		panic(fmt.Sprintf("%#v", c.conn))
	}
}
