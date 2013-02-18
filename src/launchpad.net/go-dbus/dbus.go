package dbus

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"sync"
	"sync/atomic"
)

type StandardBus int

const (
	SessionBus StandardBus = iota
	SystemBus
)

const (
	BUS_DAEMON_NAME  = "org.freedesktop.DBus"
	BUS_DAEMON_PATH  = ObjectPath("/org/freedesktop/DBus")
	BUS_DAEMON_IFACE = "org.freedesktop.DBus"
)

type MessageFilter struct {
	filter func(*Message) *Message
}

type Connection struct {
	UniqueName         string
	conn               net.Conn
	busProxy           BusDaemon
	lastSerial         uint32

	handlerMutex       sync.Mutex // covers the next three
	messageFilters     []*MessageFilter
	methodCallReplies  map[uint32] chan<- *Message
	objectPathHandlers map[ObjectPath] chan<- *Message
	signalMatchRules   signalWatchSet

	nameInfoMutex     sync.Mutex
	nameInfo          map[string] *nameInfo
}

type ObjectProxy struct {
	bus *Connection
	destination string
	path ObjectPath
}

func (o *ObjectProxy) GetObjectPath() ObjectPath {
	return o.path
}

func (o *ObjectProxy) Call(iface, method string, args ...interface{}) (*Message, error) {
	msg := NewMethodCallMessage(o.destination, o.path, iface, method)
	if err := msg.AppendArgs(args...); err != nil {
		return nil, err
	}
	reply, err := o.bus.SendWithReply(msg)
	if err != nil {
		return nil, err
	}
	if reply.Type == TypeError {
		return nil, reply.AsError()
	}
	return reply, nil
}

func (o *ObjectProxy) WatchSignal(iface, member string, handler func(*Message)) (*SignalWatch, error) {
	return o.bus.WatchSignal(&MatchRule{
		Type: TypeSignal,
		Sender: o.destination,
		Path: o.path,
		Interface: iface,
		Member: member}, handler)
}

func Connect(busType StandardBus) (*Connection, error) {
	var address string

	switch busType {
	case SessionBus:
		address = os.Getenv("DBUS_SESSION_BUS_ADDRESS")

	case SystemBus:
		if address = os.Getenv("DBUS_SYSTEM_BUS_ADDRESS"); len(address) == 0 {
			address = "unix:path=/var/run/dbus/system_bus_socket"
		}

	default:
		return nil, errors.New("Unknown bus")
	}

	trans, err := newTransport(address)
	if err != nil {
		return nil, err
	}
	bus := new(Connection)
	if bus.conn, err = trans.Dial(); err != nil {
		return nil, err
	}

	bus.busProxy = BusDaemon{bus.Object(BUS_DAEMON_NAME, BUS_DAEMON_PATH)}

	bus.messageFilters = []*MessageFilter{}
	bus.methodCallReplies = make(map[uint32] chan<- *Message)
	bus.objectPathHandlers = make(map[ObjectPath] chan<- *Message)
	bus.signalMatchRules = make(signalWatchSet)

	bus.nameInfo = make(map[string] *nameInfo)

	return bus, nil
}

func (p *Connection) Authenticate() (err error) {
	if err = authenticate(p.conn, nil); err != nil {
		return
	}
	go p._RunLoop()
	p.UniqueName, err = p.busProxy.Hello()
	return
}

func (p *Connection) _MessageReceiver(msgChan chan<- *Message) {
	for {
		msg, err := readMessage(p.conn)
		if err != nil {
			if err != io.EOF {
				log.Println("Failed to read message:", err)
			}
			break
		}
		msgChan <- msg
	}
	close(msgChan)
}

func (p *Connection) _RunLoop() {
	msgChan := make(chan *Message)
	go p._MessageReceiver(msgChan)
	for msg := range msgChan {
		p._MessageDispatch(msg)
	}
}

func (p *Connection) _MessageDispatch(msg *Message) {
	// Run the message through the registered filters, stopping
	// processing if a filter returns nil.
	for _, filter := range p.messageFilters {
		msg := filter.filter(msg)
		if msg == nil {
			return
		}
	}

	switch msg.Type {
	case TypeMethodCall:
		switch {
		case msg.Iface == "org.freedesktop.DBus.Peer" && msg.Member == "Ping":
			reply := NewMethodReturnMessage(msg)
			_ = p.Send(reply)
		case msg.Iface == "org.freedesktop.DBus.Peer" && msg.Member == "GetMachineId":
			// Should be returning the UUID found in /var/lib/dbus/machine-id
			fmt.Println("XXX: handle GetMachineId")
			reply := NewMethodReturnMessage(msg)
			_ = reply.AppendArgs("machine-id")
			_ = p.Send(reply)
		default:
			// XXX: need to lock the map
			p.handlerMutex.Lock()
			handler, ok := p.objectPathHandlers[msg.Path]
			p.handlerMutex.Unlock()
			if ok {
				handler <- msg
			} else {
				reply := NewErrorMessage(msg, "org.freedesktop.DBus.Error.UnknownObject", "Unknown object path " + string(msg.Path))
				_ = p.Send(reply)
			}
		}
	case TypeMethodReturn, TypeError:
		p.handlerMutex.Lock()
		rs := msg.replySerial
		replyChan, ok := p.methodCallReplies[rs]
		if ok {
			delete(p.methodCallReplies, rs)
		}
		p.handlerMutex.Unlock()
		if ok {
			replyChan <- msg
		}
	case TypeSignal:
		p.handlerMutex.Lock()
		watches := p.signalMatchRules.FindMatches(msg)
		p.handlerMutex.Unlock()
		for _, watch := range watches {
			watch.handler(msg)
		}
	}
}

func (p *Connection) Close() error {
	return p.conn.Close()
}

func (p *Connection) nextSerial() uint32 {
	return atomic.AddUint32(&p.lastSerial, 1)
}

func (p *Connection) Send(msg *Message) error {
	msg.setSerial(p.nextSerial())
	if _, err := msg.WriteTo(p.conn); err != nil {
		return err
	}
	return nil
}

func (p *Connection) SendWithReply(msg *Message) (*Message, error) {
	// XXX: also check for "no reply" flag.
	if msg.Type != TypeMethodCall {
		panic("Only method calls have replies")
	}
	serial := p.nextSerial()
	msg.setSerial(serial)

	replyChan := make(chan *Message, 1)
	p.handlerMutex.Lock()
	p.methodCallReplies[serial] = replyChan
	p.handlerMutex.Unlock()

	if _, err := msg.WriteTo(p.conn); err != nil {
		p.handlerMutex.Lock()
		delete(p.methodCallReplies, serial)
		p.handlerMutex.Unlock()
		return nil, err
	}

	reply := <-replyChan
	return reply, nil
}

func (p *Connection) RegisterMessageFilter(filter func (*Message) *Message) *MessageFilter {
	msgFilter := &MessageFilter{filter}
	p.messageFilters = append(p.messageFilters, msgFilter)
	return msgFilter
}

func (p *Connection) UnregisterMessageFilter(filter *MessageFilter) {
	for i, other := range p.messageFilters {
		if other == filter {
			p.messageFilters = append(p.messageFilters[:i], p.messageFilters[i+1:]...)
			return
		}
	}
	panic("Message filter not registered to this bus")
}

func (p *Connection) RegisterObjectPath(path ObjectPath, handler chan<- *Message) {
	p.handlerMutex.Lock()
	defer p.handlerMutex.Unlock()
	if _, ok := p.objectPathHandlers[path]; ok {
		panic("A handler has already been registered for " + string(path))
	}
	p.objectPathHandlers[path] = handler
}

func (p *Connection) UnregisterObjectPath(path ObjectPath) {
	p.handlerMutex.Lock()
	defer p.handlerMutex.Unlock()
	if _, ok := p.objectPathHandlers[path]; !ok {
		panic("No handler registered for " + string(path))
	}
	delete(p.objectPathHandlers, path)
}

func (p *Connection) _GetIntrospect(dest string, path ObjectPath) Introspect {
	msg := NewMethodCallMessage(dest, path, "org.freedesktop.DBus.Introspectable", "Introspect")

	reply, err := p.SendWithReply(msg)
	if err != nil {
		return nil
	}
	if v, ok := reply.GetAllArgs()[0].(string); ok {
		if intro, err := NewIntrospect(v); err == nil {
			return intro
		}
	}
	return nil
}

// Retrieve a specified object.
func (p *Connection) Object(dest string, path ObjectPath) *ObjectProxy {
	return &ObjectProxy{p, dest, path}
}
