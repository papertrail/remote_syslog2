package main

import (
	"github.com/sevenscale/remote_syslog2/syslog"
	"log"
	"time"
)

func main() {
	c, err := syslog.Dial("tcp", "localhost:1234", nil)
	if err != nil {
		log.Fatalf("failed to dial: %v", err)
	}

	p := syslog.Packet{
		Severity: syslog.SevCrit,
		Facility: syslog.LogAuth,
		Time:     time.Now(),
		Hostname: "mymachine.example.com",
		Tag:      "su",
		Message:  "hello",
	}
	log.Println(c.WritePacket(p))
}
