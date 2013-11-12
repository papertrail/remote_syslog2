package syslog

import (
	"bytes"
	"testing"
	"time"
)

func TestPacketPriority(t *testing.T) {
	tests := []struct {
		severity Priority
		facility Priority
		priority Priority
	}{
		{0, 0, 0},
		{SevNotice, LogLocal4, 165},
	}
	for _, test := range tests {
		p := Packet{Severity: test.severity, Facility: test.facility}
		if result := p.Priority(); result != test.priority {
			t.Errorf("Bad priority, got %s expected %d", result, test.priority)
		}
	}
}

func parseTime(s string) time.Time {
	t, err := time.Parse(time.RFC3339Nano, s)
	if err != nil {
		panic(err)
	}
	return t
}

func TestPacketWriteTo(t *testing.T) {
	tests := []struct {
		packet Packet
		output string
	}{
		{
			// from https://tools.ietf.org/html/rfc5424#section-6.5
			// without a MSGID
			Packet{
				Severity: SevCrit,
				Facility: LogAuth,
				Time:     parseTime("2003-10-11T22:14:15.003Z"),
				Hostname: "mymachine.example.com",
				Tag:      "su",
				Message:  "'su root' failed for lonvick on /dev/pts/8",
			},
			"<34>1 2003-10-11T22:14:15.003Z mymachine.example.com su - - - 'su root' failed for lonvick on /dev/pts/8\n",
		},
		{
			// from https://tools.ietf.org/html/rfc5424#section-6.5
			Packet{
				Severity: SevNotice,
				Facility: LogLocal4,
				Time:     parseTime("2003-08-24T05:14:15.000003-07:00"),
				Hostname: "192.0.2.1",
				Tag:      "myproc",
				Message:  `%% It's time to make the do-nuts.`,
			},
			"<165>1 2003-08-24T05:14:15.000003-07:00 192.0.2.1 myproc - - - %% It's time to make the do-nuts.\n",
		},
	}
	for _, test := range tests {
		b := new(bytes.Buffer)
		n, err := test.packet.WriteTo(b)
		if err != nil {
			t.Errorf("Unexpected error %v", err)
		}
		if int(n) != len(test.output) {
			t.Errorf("Unexpected count, expected %d got %d", len(test.output), n)
		}
		if out := b.String(); out != test.output {
			t.Errorf("Unexpected output, expected\n%v\ngot\n%v", test.output, out)
		}
	}
}
