package syslog

import (
	"testing"
)

func TestLookupSeverity(t *testing.T) {
	var sev Priority
	var err error

	sev, err = Severity("warn")
	if sev != SevWarning && err != nil {
		t.Errorf("Failed to lookup severity warning")
	}

	sev, err = Severity("foo")
	if sev != 0 && err != ErrPriority {
		t.Errorf("Failed to lookup severity foo")
	}

	sev, err = Severity("")
	if sev != 0 && err != ErrPriority {
		t.Errorf("Failed to lookup empty severity")
	}
}

func TestLookupFacility(t *testing.T) {
	var facility Priority
	var err error

	facility, err = Facility("local1")
	if facility != LogLocal1 && err != nil {
		t.Errorf("Failed to lookup facility local1")
	}

	facility, err = Facility("foo")
	if facility != 0 && err != ErrPriority {
		t.Errorf("Failed to lookup facility foo")
	}

	facility, err = Facility("")
	if facility != 0 && err != ErrPriority {
		t.Errorf("Failed to lookup empty facility")
	}
}
