package dbus

import (
	. "launchpad.net/gocheck"
	"testing"
)

func TestAll(t *testing.T) {
	TestingT(t)
}

type S struct {}

var _ = Suite(&S{})
