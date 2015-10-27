package syslog

import (
	"testing"
	"time"

	. "gopkg.in/check.v1"
)

func Test(t *testing.T) { TestingT(t) }

type PacketSuite struct {
}

var _ = Suite(&PacketSuite{})

func (s *PacketSuite) SetUpSuite(c *C) {
}

func (s *PacketSuite) TestPriority1(c *C) {
	p := &Packet{Severity: 0, Facility: 0}
	c.Assert(p.Priority(), Equals, Priority(0))
}

func (s *PacketSuite) TestPriority2(c *C) {
	p := &Packet{Severity: SevNotice, Facility: LogLocal4}
	c.Assert(p.Priority(), Equals, Priority(165))
}

func (s *PacketSuite) TestPacketGenerate(c *C) {
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
				Time:     s.time("2003-10-11T22:14:15.003Z", c),
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
				Time:     s.time("2003-08-24T05:14:15.000003-07:00", c),
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
				Time:     s.time("2003-08-24T05:14:15.123456789-07:00", c),
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
				Time:     s.time("2003-08-24T05:14:15.000003-07:00", c),
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
				Time:     s.time("2003-08-24T05:14:15.000003-07:00", c),
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
				Time:     s.time("2003-08-24T05:14:15.000003-07:00", c),
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
		c.Assert(out, Equals, test.output)
	}
}

func (s *PacketSuite) time(v string, c *C) time.Time {
	t, err := time.Parse(time.RFC3339Nano, v)
	c.Assert(err, IsNil)
	return t
}
