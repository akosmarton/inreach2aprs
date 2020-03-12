package main

import (
	"bytes"
	"fmt"
	"math"
	"net"
	"strings"
	"time"
)

type AprsClient struct {
	Host     string
	Port     int
	User     string
	Passcode string
}

type AprsPacket struct {
	Callsign  string
	Timestamp time.Time
	Latitude  float64
	Longitude float64
	Symbol    string
	Course    int
	Speed     int
	Altitude  int
	Comment   string
}

func NewAprsClient(host string, port int, user string, passcode string) *AprsClient {
	return &AprsClient{
		Host:     host,
		Port:     port,
		User:     user,
		Passcode: passcode,
	}
}

func (c *AprsClient) Send(p *AprsPacket) error {
	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", c.Host, c.Port))
	if err != nil {
		return err
	}
	if _, err := conn.Write([]byte(fmt.Sprintf("user %s pass %s\n", c.User, c.Passcode))); err != nil {
		return err
	}
	if _, err := conn.Write(p.Encode()); err != nil {
		return err
	}
	return nil
}

func (p *AprsPacket) Encode() []byte {
	b := &bytes.Buffer{}
	b.WriteString(fmt.Sprintf("%s>APRS,TCPIP*:", strings.ToUpper(p.Callsign)))
	if p.Timestamp.IsZero() {
		b.WriteString("!")
	} else {
		b.WriteString("/")
		b.WriteString(p.Timestamp.UTC().Format("021504z"))
	}

	if p.Latitude != 0 || p.Longitude != 0 {
		var latDir, longDir string
		if p.Latitude > 0 {
			latDir = "N"
		} else {
			latDir = "S"
		}
		if p.Longitude > 0 {
			longDir = "E"
		} else {
			longDir = "W"
		}
		b.WriteString(fmt.Sprintf("%02.0f%02.2f%s/%03.0f%02.2f%s", math.Trunc(p.Latitude), (p.Latitude-math.Trunc(p.Latitude))*60, latDir, math.Trunc(p.Longitude), (p.Longitude-math.Trunc(p.Longitude))*60, longDir))
	}

	b.WriteString(p.Symbol)

	if p.Course != 0 || p.Speed != 0 {
		b.WriteString(fmt.Sprintf("%03d/%03d", p.Course, p.Speed))
	}

	if p.Altitude != 0 {
		b.WriteString(fmt.Sprintf("/A=%06d", p.Altitude))
	}

	b.WriteString(p.Comment)

	b.WriteString("\n")

	return b.Bytes()
}
