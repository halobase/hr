package hr

import (
	"net"
	"time"
)

type Type interface {
	Parse(string) error
}

type Time time.Time

func (t *Time) Parse(s string) error {
	p, err := time.Parse(time.RFC3339, s)
	if err != nil {
		return err
	}
	*t = Time(p)
	return nil
}

func (t Time) MarshalText() ([]byte, error) {
	return time.Time(t).MarshalText()
}

type Duration time.Duration

func (d *Duration) Parse(s string) error {
	p, err := time.ParseDuration(s)
	if err != nil {
		return err
	}
	*d = Duration(p)
	return nil
}

type IP net.IP

func (ip *IP) Parse(s string) error {
	*ip = IP(net.ParseIP(s))
	return nil
}
