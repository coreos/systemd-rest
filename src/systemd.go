package main

import "github.com/philips/go-dbus"
import "log"

func main() {
	var (
		err    error
		conn   *dbus.Connection
		method *dbus.Method
		out    []interface{}
	)

	// Connect to Session or System buses.
	if conn, err = dbus.Connect(dbus.SystemBus); err != nil {
		log.Fatal("Connection error:", err)
	}
	if err = conn.Authenticate(); err != nil {
		log.Fatal("Authentication error:", err)
	}

	// Get objects.
	obj := conn.Object("org.freedesktop.systemd1", "/org/freedesktop/systemd1")
	if err != nil {
		log.Fatal("Object error:", err)
	}

	// Introspect objects.
	var intro dbus.Introspect
	method, err = obj.Interface().Method("Introspect")
	if err != nil {
		log.Fatal("Interface error:", err)
	}
	out, err = conn.Call("org.freedesktop.DBus.Introspectable", "Introspect", method)
	if err != nil {
		log.Fatal("Introspect error:", err)
	}
	intro, err = dbus.NewIntrospect(out[0].(string))
	m := intro.GetInterfaceData("org.freedesktop.systemd1.Manager").GetMethodData("StartUnit")
	log.Printf("%s in:%s out:%s", m.GetName(), m.GetInSignature(), m.GetOutSignature())

	// Call object methods.
	method, err = obj.Interface("org.freedesktop.systemd1.Manager").Method("StartUnit")
	if err != nil {
		log.Fatal(err)
	}
	out, err = conn.Call(method, "acpid.service", "fail")
	if err != nil {
		log.Fatal("Notification error:", err)
	}
	log.Print("Notification id:", out[0])
}
