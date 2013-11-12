package syslog

import (
	"crypto/tls"
	"fmt"
	"github.com/sevenscale/remote_syslog2/syslog/certs"
	"net"
)

type Conn struct {
	conn net.Conn
}

func Dial(network, raddr string, bundle *certs.CertBundle) (*Conn, error) {
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
		// todo: store hostname?
		return &Conn{conn}, nil
	}
}

func (c Conn) WritePacket(p Packet) error {
	_, err := p.WriteTo(c.conn)
	return err
}
