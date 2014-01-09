package syslog

import (
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

func TestPacketGenerate(t *testing.T) {
	tests := []struct {
		packet   Packet
		max_size int
		output   string
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
			0,
			"<34>1 2003-10-11T22:14:15.003Z mymachine.example.com su - - - 'su root' failed for lonvick on /dev/pts/8",
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
			0,
			"<165>1 2003-08-24T05:14:15.000003-07:00 192.0.2.1 myproc - - - %% It's time to make the do-nuts.",
		},
		{
			// test that fractional seconds is at most 6 digits long
			Packet{
				Severity: SevNotice,
				Facility: LogLocal4,
				Time:     parseTime("2003-08-24T05:14:15.123456789-07:00"),
				Hostname: "192.0.2.1",
				Tag:      "myproc",
				Message:  `%% It's time to make the do-nuts.`,
			},
			0,
			"<165>1 2003-08-24T05:14:15.123456-07:00 192.0.2.1 myproc - - - %% It's time to make the do-nuts.",
		},
		{
			// test truncation
			Packet{
				Severity: SevNotice,
				Facility: LogLocal4,
				Time:     parseTime("2003-08-24T05:14:15.000003-07:00"),
				Hostname: "192.0.2.1",
				Tag:      "myproc",
				Message:  `%% It's time to make the do-nuts.`,
			},
			75,
			"<165>1 2003-08-24T05:14:15.000003-07:00 192.0.2.1 myproc - - - %% It's time",
		},
		{
			// test truncation isn't applied when message is already short enough
			Packet{
				Severity: SevNotice,
				Facility: LogLocal4,
				Time:     parseTime("2003-08-24T05:14:15.000003-07:00"),
				Hostname: "192.0.2.1",
				Tag:      "myproc",
				Message:  `%% It's time to make the do-nuts.`,
			},
			97,
			"<165>1 2003-08-24T05:14:15.000003-07:00 192.0.2.1 myproc - - - %% It's time to make the do-nuts.",
		},
		{
			Packet{
				Severity: SevNotice,
				Facility: LogLocal4,
				Time:     parseTime("2003-08-24T05:14:15.000003-07:00"),
				Hostname: "192.0.2.1",
				Tag:      "myproc",
				Message:  "newline:'\n'. nullbyte:'\x00'. carriage return:'\r'.",
			},
			0,
			"<165>1 2003-08-24T05:14:15.000003-07:00 192.0.2.1 myproc - - - newline:' '. nullbyte:' '. carriage return:' '.",
		},
	}
	for _, test := range tests {
		out := test.packet.Generate(test.max_size)
		if out != test.output {
			t.Errorf("Unexpected output, expected\n%v\ngot\n%v", test.output, out)
		}
	}
}
