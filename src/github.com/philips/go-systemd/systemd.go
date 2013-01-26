package systemd

import "github.com/philips/go-dbus"
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

func (s *Systemd1) StartUnit(name string, mode string) (ret interface{}, err error) {
	var (
		method *dbus.Method
		out    []interface{}
	)

	obj := s.conn.Object("org.freedesktop.systemd1", "/org/freedesktop/systemd1")

	method, err = obj.Interface("org.freedesktop.systemd1.Manager").Method("StartUnit")
	if err != nil {
		return nil, err
	}

	out, err = s.conn.Call(method, name, mode)
	if err != nil {
		return nil, err
	}

	return out[0], err
}
