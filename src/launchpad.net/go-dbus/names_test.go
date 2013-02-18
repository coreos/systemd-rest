package dbus

import (
	. "launchpad.net/gocheck"
)

func (s *S) TestConnectionWatchName(c *C) {
	bus, err := Connect(SessionBus)
	c.Assert(err, Equals, nil)
	defer bus.Close()
	c.Assert(bus.Authenticate(), Equals, nil)

	// Set up the name watch
	nameChanged := make(chan int, 1)
	owners := []string{}
	watch, err := bus.WatchName("com.example.GoDbus", func (newOwner string) {
		owners = append(owners, newOwner)
		nameChanged <- 0
	})
	c.Assert(err, Equals, nil)
	defer watch.Cancel()

	// Our handler will be called once with the initial name owner
	<- nameChanged
	c.Check(owners, DeepEquals, []string{""})

	// Acquire the name, and wait for the process to complete.
	nameAcquired := make(chan int, 1)
	name := bus.RequestName("com.example.GoDbus", NameFlagDoNotQueue, func(*BusName) { nameAcquired <- 0 }, nil)
	<- nameAcquired

	<- nameChanged
	c.Check(owners, DeepEquals, []string{"", bus.UniqueName})

	err = name.Release()
	c.Assert(err, Equals, nil)
	<- nameChanged
	c.Check(owners, DeepEquals, []string{"", bus.UniqueName, ""})
}

func (s *S) TestConnectionRequestName(c *C) {
	bus, err := Connect(SessionBus)
	c.Assert(err, Equals, nil)
	defer bus.Close()
	c.Assert(bus.Authenticate(), Equals, nil)

	nameAcquired := make(chan int, 1)
	name := bus.RequestName("com.example.GoDbus", 0, func (*BusName) { nameAcquired <- 0 }, nil)
	c.Check(name, Not(Equals), nil)

	<- nameAcquired
	owner, err := bus.busProxy.GetNameOwner("com.example.GoDbus")
	c.Check(err, Equals, nil)
	c.Check(owner, Equals, bus.UniqueName)

	c.Check(name.Release(), Equals, nil)
}

func (s *S) TestConnectionRequestNameQueued(c *C) {
	// Acquire the name on a second connection
	bus1, err := Connect(SessionBus)
	c.Assert(err, Equals, nil)
	defer bus1.Close()
	c.Assert(bus1.Authenticate(), Equals, nil)

	bus2, err := Connect(SessionBus)
	c.Assert(err, Equals, nil)
	defer bus2.Close()
	c.Assert(bus2.Authenticate(), Equals, nil)

	ready := make(chan int, 1)
	name1 := bus1.RequestName("com.example.GoDbus", 0, func (*BusName) { ready <- 0 }, nil)
	<- ready
	c.Check(name1.needsRelease, Equals, true)

	callLog := []string{}
	called := make(chan int, 1)
	name2 := bus2.RequestName("com.example.GoDbus", 0,
		func (*BusName) {
			callLog = append(callLog, "acquired")
			called <- 0
		}, func(*BusName) {
			callLog = append(callLog, "lost")
			called <- 0
		})
	<- called
	c.Check(name2.needsRelease, Equals, true)
	c.Check(callLog, DeepEquals, []string{"lost"})

	// Release the name on the first connection
	c.Check(name1.Release(), Equals, nil)

	<- called
	c.Check(callLog, DeepEquals, []string{"lost", "acquired"})
	c.Check(name2.Release(), Equals, nil)
}

func (s *S) TestConnectionRequestNameDoNotQueue(c *C) {
	// Acquire the name on a second connection
	bus1, err := Connect(SessionBus)
	c.Assert(err, Equals, nil)
	defer bus1.Close()
	c.Assert(bus1.Authenticate(), Equals, nil)

	bus2, err := Connect(SessionBus)
	c.Assert(err, Equals, nil)
	defer bus2.Close()
	c.Assert(bus2.Authenticate(), Equals, nil)

	ready := make(chan int, 1)
	name1 := bus1.RequestName("com.example.GoDbus", 0, func (*BusName) { ready <- 0 }, nil)
	defer name1.Release()
	<- ready
	c.Check(name1.needsRelease, Equals, true)

	callLog := []string{}
	called := make(chan int, 1)
	name2 := bus2.RequestName("com.example.GoDbus", NameFlagDoNotQueue,
		func (*BusName) {
			callLog = append(callLog, "acquired")
			called <- 0
		}, func(*BusName) {
			callLog = append(callLog, "lost")
			called <- 0
		})
	<- called
	c.Check(name2.needsRelease, Equals, false)
	c.Check(callLog, DeepEquals, []string{"lost"})

	c.Check(name2.Release(), Equals, nil)
}

func (s *S) TestConnectionRequestNameAllowReplacement(c *C) {
	// Acquire the name on a second connection
	bus1, err := Connect(SessionBus)
	c.Assert(err, Equals, nil)
	defer bus1.Close()
	c.Assert(bus1.Authenticate(), Equals, nil)

	bus2, err := Connect(SessionBus)
	c.Assert(err, Equals, nil)
	defer bus2.Close()
	c.Assert(bus2.Authenticate(), Equals, nil)

	callLog1 := []string{}
	called1 := make(chan int, 1)
	name1 := bus1.RequestName("com.example.GoDbus", NameFlagAllowReplacement,
		func (*BusName) {
			callLog1 = append(callLog1, "acquired")
			called1 <- 0
		}, func(*BusName) {
			callLog1 = append(callLog1, "lost")
			called1 <- 0
		})
	defer name1.Release()
	<- called1
	c.Check(name1.needsRelease, Equals, true)
	c.Check(callLog1, DeepEquals, []string{"acquired"})

	callLog2 := []string{}
	called2 := make(chan int, 1)
	name2 := bus2.RequestName("com.example.GoDbus", NameFlagReplaceExisting,
		func (*BusName) {
			callLog2 = append(callLog2, "acquired")
			called2 <- 0
		}, func(*BusName) {
			callLog2 = append(callLog2, "lost")
			called2 <- 0
		})
	defer name2.Release()
	<- called2
	c.Check(name2.needsRelease, Equals, true)
	c.Check(callLog2, DeepEquals, []string{"acquired"})

	// The first name owner loses possession.
	<- called1
	c.Check(callLog1, DeepEquals, []string{"acquired", "lost"})
}
