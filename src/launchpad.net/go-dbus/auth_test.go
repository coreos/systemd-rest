package dbus

import (
	"bufio"
	"net"
	. "launchpad.net/gocheck"
)

func (s *S) TestAuthenticate(c *C) {
	server, client := net.Pipe()
	clientWrites := []string{}
	complete := make(chan int)
	go func() {
		r := bufio.NewReader(server)
		// Read the nul byte that marks the start of the protocol
		zero := []byte{0}
		r.Read(zero)

		clientWrites = append(clientWrites, string(zero))
		line, _, _ := r.ReadLine()
		clientWrites = append(clientWrites, string(line))

		server.Write([]byte("OK\r\n"))
		line, _, _ = r.ReadLine()
		clientWrites = append(clientWrites, string(line))

		complete <- 1
	}()

	c.Check(authenticate(client, nil), Equals, nil)
	<- complete
	c.Check(clientWrites[0], Equals, "\x00")
	c.Check(clientWrites[1][:13], Equals, "AUTH EXTERNAL")
	c.Check(clientWrites[2], Equals, "BEGIN")
}
