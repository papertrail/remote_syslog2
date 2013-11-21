package syslog

import (
	"crypto/tls"
	"fmt"
	"github.com/howbazaar/loggo"
	"github.com/sevenscale/remote_syslog2/syslog/certs"
	"net"
)

var log = loggo.GetLogger("syslog")

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

func (c Conn) WritePacket(p Packet) error {
	_, err := p.WriteTo(c.conn)
	return err
}
