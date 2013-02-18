package dbus

import (
	. "launchpad.net/gocheck"
)

func (s *S) TestConnectionWatchSignal(c *C) {
	bus1, err := Connect(SessionBus)
	c.Assert(err, Equals, nil)
	defer bus1.Close()
	c.Assert(bus1.Authenticate(), Equals, nil)

	// Set up a second bus connection to receive a signal.
	watchReady := make(chan int)
	complete := make(chan *Message)
	go func(sender string, watchReady chan<- int, complete chan<- *Message) {
		bus2, err := Connect(SessionBus)
		if err != nil {
			c.Error(err)
			watchReady <- 0
			complete <- nil
			return
		}
		defer bus2.Close()
		if err := bus2.Authenticate(); err != nil {
			c.Error(err)
			watchReady <- 0
			complete <- nil
			return
		}
		msgChan := make(chan *Message)
		watch, err := bus2.WatchSignal(&MatchRule{
			Type: TypeSignal,
			Sender: sender,
			Path: "/go/dbus/test",
			Interface: "com.example.GoDbus",
			Member: "TestSignal"},
			func(msg *Message) { msgChan <- msg })
		watchReady <- 0
		if err != nil {
			c.Error(err)
			bus2.Close()
			complete <- nil
			return
		}
		msg := <-msgChan
		if err := watch.Cancel(); err != nil {
			c.Error(err)
		}
		complete <- msg
	}(bus1.UniqueName, watchReady, complete)

	// Wait for the goroutine to configure the signal watch
	<-watchReady

	// Send the signal and wait for it to be received at the other end.
	signal := NewSignalMessage("/go/dbus/test", "com.example.GoDbus", "TestSignal")
	if err := bus1.Send(signal); err != nil {
		c.Fatal(err)
	}

	signal2 := <- complete
	c.Check(signal2, Not(Equals), nil)
}

func (s *S) TestConnectionWatchSignalWithBusName(c *C) {
	bus, err := Connect(SessionBus)
	c.Assert(err, Equals, nil)
	defer bus.Close()
	c.Assert(bus.Authenticate(), Equals, nil)

	// Request a bus name
	result, err := bus.busProxy.RequestName("com.example.GoDbus", 0x4)
	c.Assert(err, Equals, nil)
	c.Assert(result, Equals, uint32(1)) // We are Primary Owner

	// Set up a signal watch
	received := make(chan *Message, 1)
	watch, err := bus.WatchSignal(&MatchRule{
		Type: TypeSignal,
		Sender: "com.example.GoDbus",
		Interface: "com.example.GoDbus",
		Member: "TestSignal"},
		func(msg *Message) { received <- msg })
	c.Assert(err, Equals, nil)
	defer watch.Cancel()

	// Send the signal, and wait to receive it.
	signal := NewSignalMessage("/go/dbus/test", "com.example.GoDbus", "TestSignal")
	if err := bus.Send(signal); err != nil {
		c.Fatal(err)
	}

	signal2 := <- received
	c.Check(signal2, Not(Equals), nil)
}

func (s *S) TestSignalWatchSetAdd(c *C) {
	set := make(signalWatchSet)
	watch := SignalWatch{rule: MatchRule{
		Type: TypeSignal,
		Sender: ":1.42",
		Path: "/foo",
		Interface: "com.example.Foo",
		Member: "Bar"}}
	set.Add(&watch)

	byInterface, ok := set["/foo"]
	c.Assert(ok, Equals, true)
	byMember, ok := byInterface["com.example.Foo"]
	c.Assert(ok, Equals, true)
	watches, ok := byMember["Bar"]
	c.Assert(ok, Equals, true)
	c.Check(watches, DeepEquals, []*SignalWatch{&watch})
}

func (s *S) TestSignalWatchSetRemove(c *C) {
	set := make(signalWatchSet)
	watch1 := SignalWatch{rule: MatchRule{
		Type: TypeSignal,
		Sender: ":1.42",
		Path: "/foo",
		Interface: "com.example.Foo",
		Member: "Bar"}}
	set.Add(&watch1)
	watch2 := SignalWatch{rule: MatchRule{
		Type: TypeSignal,
		Sender: ":1.43",
		Path: "/foo",
		Interface: "com.example.Foo",
		Member: "Bar"}}
	set.Add(&watch2)

	c.Check(set.Remove(&watch1), Equals, true)
	c.Check(set["/foo"]["com.example.Foo"]["Bar"], DeepEquals, []*SignalWatch{&watch2})

	// A second attempt at removal fails
	c.Check(set.Remove(&watch1), Equals, false)
}

func (s *S) TestSignalWatchSetFindMatches(c *C) {
	msg := NewSignalMessage("/foo", "com.example.Foo", "Bar")
	msg.Sender = ":1.42"

	set := make(signalWatchSet)
	watch := SignalWatch{rule: MatchRule{
		Type: TypeSignal,
		Sender: ":1.42",
		Path: "/foo",
		Interface: "com.example.Foo",
		Member: "Bar"}}

	set.Add(&watch)
	c.Check(set.FindMatches(msg), DeepEquals, []*SignalWatch{&watch})
	set.Remove(&watch)

	// An empty path also matches
	watch.rule.Path = ""
	set.Add(&watch)
	c.Check(set.FindMatches(msg), DeepEquals, []*SignalWatch{&watch})
	set.Remove(&watch)

	// Or an empty interface
	watch.rule.Path = "/foo"
	watch.rule.Interface = ""
	set.Add(&watch)
	c.Check(set.FindMatches(msg), DeepEquals, []*SignalWatch{&watch})
	set.Remove(&watch)

	// Or an empty member
	watch.rule.Interface = "com.example.Foo"
	watch.rule.Member = ""
	set.Add(&watch)
	c.Check(set.FindMatches(msg), DeepEquals, []*SignalWatch{&watch})
	set.Remove(&watch)
}
