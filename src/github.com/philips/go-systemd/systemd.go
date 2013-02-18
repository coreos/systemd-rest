package systemd

import "launchpad.net/go-dbus"
type Systemd1 struct {
	conn   *dbus.Connection
}

type Job struct {
	Id string `json:"job"`
	Error string `json:"error"`
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

func (s *Systemd1) StartUnit(name string, mode string) (job Job, err error) {
	obj := s.conn.Object("org.freedesktop.systemd1", "/org/freedesktop/systemd1")

	reply, err := obj.Call("org.freedesktop.systemd1.Manager", "StartUnit",
		name, mode)
	if err != nil {
		return Job{"", err.Error()}, err
	}

	err = reply.GetArgs(&job.Id)

	return job, err
}

func (s *Systemd1) StopUnit(name string, mode string) (job Job, err error) {
	obj := s.conn.Object("org.freedesktop.systemd1", "/org/freedesktop/systemd1")

	reply, err := obj.Call("org.freedesktop.systemd1.Manager", "StopUnit",
		name, mode)
	if err != nil {
		return Job{"", err.Error()}, err
	}

	err = reply.GetArgs(&job.Id)

	return job, err
}

func (s *Systemd1) ListUnits() (message string, err error) {
	obj := s.conn.Object("org.freedesktop.systemd1", "/org/freedesktop/systemd1")

	reply, err := obj.Call("org.freedesktop.systemd1.Manager", "ListUnits")
	if err != nil {
		return "", err
	}

	err = reply.GetArgs(&message)

	return message, err
}
