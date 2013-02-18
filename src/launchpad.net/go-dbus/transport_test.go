package dbus

import . "launchpad.net/gocheck"

func (s *S) TestNewTransportUnix(c *C) {
	trans, err := newTransport("unix:path=/tmp/dbus%3dsock")
	c.Check(err, Equals, nil)
	unixTrans, ok := trans.(*unixTransport)
	c.Check(ok, Equals, true)
	c.Check(unixTrans.Address, Equals, "/tmp/dbus=sock")

	// And for abstract namespace sockets:
	trans, err = newTransport("unix:abstract=/tmp/dbus%3dsock")
	c.Check(err, Equals, nil)
	unixTrans, ok = trans.(*unixTransport)
	c.Check(ok, Equals, true)
	c.Check(unixTrans.Address, Equals, "@/tmp/dbus=sock")
}

func (s *S) TestNewTransportTcp(c *C) {
	trans, err := newTransport("tcp:host=localhost,port=4444")
	c.Check(err, Equals, nil)
	tcpTrans, ok := trans.(*tcpTransport)
	c.Check(ok, Equals, true)
	c.Check(tcpTrans.Address, Equals, "localhost:4444")
	c.Check(tcpTrans.Family, Equals, "tcp4")

	// And with explicit family:
	trans, err = newTransport("tcp:host=localhost,port=4444,family=ipv4")
	c.Check(err, Equals, nil)
	tcpTrans, ok = trans.(*tcpTransport)
	c.Check(ok, Equals, true)
	c.Check(tcpTrans.Address, Equals, "localhost:4444")
	c.Check(tcpTrans.Family, Equals, "tcp4")

	trans, err = newTransport("tcp:host=localhost,port=4444,family=ipv6")
	c.Check(err, Equals, nil)
	tcpTrans, ok = trans.(*tcpTransport)
	c.Check(ok, Equals, true)
	c.Check(tcpTrans.Address, Equals, "localhost:4444")
	c.Check(tcpTrans.Family, Equals, "tcp6")
}
