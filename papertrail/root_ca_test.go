package papertrail

import (
	"crypto/md5"
	"fmt"
	"testing"
)

func TestCerts(t *testing.T) {
	h := md5.New()
	h.Write(certs())
	expected := "3203239f3684b19dc0aec8554a2613a7"
	actual := fmt.Sprintf("%x", h.Sum(nil))
	if actual != expected {
		t.Errorf("Mismatched hash for papertrail certs, expected %s got %s", expected, actual)
	}
}

func TestRootCA(t *testing.T) {
	pool := RootCA()
	expected := 154
	actual := len(pool.Subjects())
	if actual != expected {
		t.Errorf("Error loading RootCA, expected %d subjects got %d", expected, actual)
	}
}
