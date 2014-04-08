package papertrail

import (
	"crypto/md5"
	"fmt"
	"testing"
)

func TestCerts(t *testing.T) {
	h := md5.New()
	h.Write(certs())
	expected := "cee9b8d2d503188ccecbb22b49cd3bec"
	actual := fmt.Sprintf("%x", h.Sum(nil))
	if actual != expected {
		t.Errorf("Mismatched hash for papertrail certs, expected %s got %s", expected, actual)
	}
}

func TestRootCA(t *testing.T) {
	pool := RootCA()
	if len(pool.Subjects()) != 3 {
		t.Errorf("error loading RootCA")
	}
}
