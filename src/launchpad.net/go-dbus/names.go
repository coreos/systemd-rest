package dbus

import (
	"errors"
	"log"
)

type nameInfo struct {
	bus          *Connection
	busName      string
	currentOwner string
	signalWatch  *SignalWatch
	watches      []*NameWatch
}

type NameWatch struct {
	info      *nameInfo
	handler   func(newOwner string)
	cancelled bool
}

func newNameInfo(bus *Connection, busName string) (*nameInfo, error) {
	info := &nameInfo{
		bus: bus,
		busName: busName,
		watches: []*NameWatch{}}
	handler := func(msg *Message) {
		var busName, oldOwner, newOwner string
		if err := msg.GetArgs(&busName, &oldOwner, &newOwner); err != nil {
			log.Println("Could not decode NameOwnerChanged message:", err)
			return
		}
		info.handleOwnerChange(newOwner)
	}
	watch, err := bus.WatchSignal(&MatchRule{
		Type: TypeSignal,
		Sender: BUS_DAEMON_NAME,
		Path: BUS_DAEMON_PATH,
		Interface: BUS_DAEMON_IFACE,
		Member: "NameOwnerChanged",
		Arg0: busName}, handler)
	if err != nil {
		return nil, err
	}
	info.signalWatch = watch

	// spawn a goroutine to find the current name owner
	go info.checkCurrentOwner()

	return info, nil
}

func (self *nameInfo) checkCurrentOwner() {
	currentOwner, err := self.bus.busProxy.GetNameOwner(self.busName)
	if err != nil {
		if dbusErr, ok := err.(*Error); !ok || dbusErr.Name != "org.freedesktop.DBus.Error.NameHasNoOwner" {
			log.Println("Unexpected error from GetNameOwner:", err)
		}
	}
	if self.currentOwner == "" {
		// Simulate an ownership change message.
		self.handleOwnerChange(currentOwner)
	}
}

func (self *nameInfo) handleOwnerChange(newOwner string) {
	for _, watch := range self.watches {
		if watch.handler != nil {
			watch.handler(newOwner)
		}
	}
	self.currentOwner = newOwner
}

func (p *Connection) WatchName(busName string, handler func(newOwner string)) (watch *NameWatch, err error) {
	p.nameInfoMutex.Lock()
	info, ok := p.nameInfo[busName]
	if !ok {
		if info, err = newNameInfo(p, busName); err != nil {
			p.nameInfoMutex.Unlock()
			return
		}
		p.nameInfo[busName] = info
	}
	watch = &NameWatch{info: info, handler: handler}
	info.watches = append(info.watches, watch)
	p.nameInfoMutex.Unlock()

	// If we're hooking up to an existing nameOwner and it already
	// knows the current name owner, tell our callback.
	if !ok && info.currentOwner != "" {
		handler(info.currentOwner)
	}
	return
}

func (watch *NameWatch) Cancel() error {
	if watch.cancelled {
		return nil
	}
	watch.cancelled = true

	info := watch.info
	bus := info.bus
	bus.nameInfoMutex.Lock()
	defer bus.nameInfoMutex.Unlock()

	found := false
	for i, other := range(info.watches) {
		if other == watch {
			info.watches[i] = info.watches[len(info.watches)-1]
			info.watches = info.watches[:len(info.watches)-1]
			found = true
			break
		}
	}
	if !found {
		return errors.New("NameOwnerWatch already cancelled")
	}
	if len(info.watches) != 0 {
		// There are other watches interested in this name, so
		// leave the nameOwner in place.
		return nil
	}
	delete(bus.nameInfo, info.busName)
	return info.signalWatch.Cancel()
}


type BusName struct {
	bus *Connection
	Name string
	Flags NameFlags

	cancelled bool
	needsRelease bool

	acquiredCallback func (*BusName)
	lostCallback     func(*BusName)

	acquiredWatch *SignalWatch
	lostWatch     *SignalWatch
}

type NameFlags uint32

const (
	NameFlagAllowReplacement = NameFlags(0x1)
	NameFlagReplaceExisting = NameFlags(0x2)
	NameFlagDoNotQueue = NameFlags(0x4)
)

func (p *Connection) RequestName(busName string, flags NameFlags, nameAcquired func(*BusName), nameLost func(*BusName)) *BusName {
	name := &BusName{
		bus: p,
		Name: busName,
		Flags: flags,
		acquiredCallback: nameAcquired,
		lostCallback: nameLost}
	go name.request()
	return name
}

func (name *BusName) request() {
	if name.cancelled {
		return
	}
	result, err := name.bus.busProxy.RequestName(name.Name, uint32(name.Flags))
	if err != nil {
		log.Println("Error requesting bus name", name.Name, "err =", err)
		return
	}
	subscribe := false
	switch result {
	case 1:
		// DBUS_REQUEST_NAME_REPLY_PRIMARY_OWNER
		if name.acquiredCallback != nil {
			name.acquiredCallback(name)
		}
		subscribe = true
		name.needsRelease = true
	case 2:
		// DBUS_REQUEST_NAME_REPLY_IN_QUEUE
		if name.lostCallback != nil {
			name.lostCallback(name)
		}
		subscribe = true
		name.needsRelease = true
	case 3:
		// DBUS_REQUEST_NAME_REPLY_EXISTS
		fallthrough
	case 4:
		// DBUS_REQUEST_NAME_REPLY_ALREADY_OWNER
		fallthrough
	default:
		// assume that other responses mean we couldn't own
		// the name
		if name.lostCallback != nil {
			name.lostCallback(name)
		}
	}

	if subscribe && !name.cancelled {
		watch, err := name.bus.WatchSignal(&MatchRule{
			Type: TypeSignal,
			Sender: BUS_DAEMON_NAME,
			Path: BUS_DAEMON_PATH,
			Interface: BUS_DAEMON_IFACE,
			Member: "NameLost",
			Arg0: name.Name},
			func(msg *Message) {
				if !name.cancelled && name.lostCallback != nil {
					name.lostCallback(name)
				}
			})
		if err != nil {
			log.Println("Could not set up NameLost signal watch")
			name.Release()
			return
		}
		name.lostWatch = watch

		watch, err = name.bus.WatchSignal(&MatchRule{
			Type: TypeSignal,
			Sender: BUS_DAEMON_NAME,
			Path: BUS_DAEMON_PATH,
			Interface: BUS_DAEMON_IFACE,
			Member: "NameAcquired",
			Arg0: name.Name},
			func(msg *Message) {
				if !name.cancelled && name.acquiredCallback != nil {
					name.acquiredCallback(name)
				}
			})
		if err != nil {
			log.Println("Could not set up NameLost signal watch")
			name.Release()
			return
		}
		name.acquiredWatch = watch

		// XXX: if we disconnect from the bus, we should
		// report the name being lost.
	}
}

func (name *BusName) Release() error {
	if name.cancelled {
		return nil
	}
	name.cancelled = true
	if name.acquiredWatch != nil {
		if err := name.acquiredWatch.Cancel(); err != nil {
			return err
		}
		name.acquiredWatch = nil
	}
	if name.lostWatch != nil {
		if err := name.lostWatch.Cancel(); err != nil {
			return err
		}
		name.lostWatch = nil
	}

	if name.needsRelease {
		result, err := name.bus.busProxy.ReleaseName(name.Name)
		if err != nil {
			return err
		}
		if result != 1 { // DBUS_RELEASE_NAME_REPLY_RELEASED
			log.Println("Unexpected result when releasing name", name.Name, "result =", result)
		}
		name.needsRelease = false
	}
	return nil
}
