package systemd

import "launchpad.net/go-dbus"
type Systemd1 struct {
	conn   *dbus.Connection
}

func (s *Systemd1) Connect() (err error) {
	conn, err := dbus.Connect(dbus.SystemBus)
	if err != nil {
		return err
	}

	err = conn.Authenticate()

	s.conn = conn

	return err
}

func (s *Systemd1) StartUnit(name string, mode string) (message string, err error) {
	obj := s.conn.Object("org.freedesktop.systemd1", "/org/freedesktop/systemd1")

	reply, err := obj.Call("org.freedesktop.systemd1.Manager", "StartUnit",
		name, mode)
	if err != nil {
		return "", err
	}

	err = reply.GetArgs(&message)

	return message, err
}
