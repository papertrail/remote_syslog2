package certs

import (
	"crypto/x509"
	"errors"
	"io/ioutil"
	"os"
)

type CertBundle struct {
	CertPool x509.CertPool
}

func NewCertBundle() CertBundle {
	return CertBundle{CertPool: *x509.NewCertPool()}
}

func (c *CertBundle) ImportFromFile(pemfile string) error {
	data, err := ioutil.ReadFile(pemfile)
	if err == nil {
		err = c.ImportBytes(data)
		if err != nil {
			return err
		}
	} else {
		return err
	}

	return nil
}

func (c *CertBundle) ImportBytes(bytes []byte) error {
	ok := c.CertPool.AppendCertsFromPEM(bytes)

	if ok != true {
		errors.New("Failed to import PEM data to certificate")
	}

	return nil
}

func (c *CertBundle) ImportFromFiles(pemfiles []string) error {
	for _, file := range pemfiles {
		err := c.ImportFromFile(file)
		if err != nil {
			return err
		}
	}
	return nil
}

func (c *CertBundle) LoadDefaultBundle() (string, error) {

	defaultCertLocations := []string{
		"/etc/remote_syslog/ca-bundle.crt",       // Local CA File for remote_syslog
		"/etc/ssl/certs/ca-certificates.crt",     // Linux etc
		"/etc/pki/tls/certs/ca-bundle.crt",       // Fedora/RHEL
		"/etc/ssl/ca-bundle.pem",                 // OpenSUSE
		"/etc/ssl/cert.pem",                      // OpenBSD
		"/usr/local/share/certs/ca-root-nss.crt", // FreeBSD
		"/usr/local/share/ca-bundle.crt",         //OS X with Brew
	}

	for _, file := range defaultCertLocations {
		if _, err := os.Stat(file); !os.IsNotExist(err) {
			err := c.ImportFromFile(file)
			if err != nil {
				return "", err
			} else {
				return file, nil
			}
		}
	}
	return "", errors.New("Unable to locate a default ca bundle")
}
