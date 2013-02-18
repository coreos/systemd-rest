Documentation
=============

Look at the API on [GoPkgDoc](http://gopkgdoc.appspot.com/pkg/github.com/norisatir/go-dbus).

Installation
============

    go get launchpad.net/~jamesh/go-dbus/trunk

Usage
=====

An example
----------

```go
// Issue OSD notifications according to the Desktop Notifications Specification 1.1
//      http://people.canonical.com/~agateau/notifications-1.1/spec/index.html
// See also
//      https://wiki.ubuntu.com/NotifyOSD#org.freedesktop.Notifications.Notify
package main

import "launchpad.net/~jamesh/go-dbus/trunk"
import "log"

func main() {
    var (
        err error
        conn *dbus.Connection
    )

    // Connect to Session or System buses.
    if conn, err = dbus.Connect(dbus.SessionBus); err != nil {
        log.Fatal("Connection error:", err)
    }
    if err = conn.Authenticate(); err != nil {
        log.Fatal("Authentication error:", err)
    }

    // Create an object proxy
    obj := conn.Object("org.freedesktop.Notifications", "/org/freedesktop/Notifications")

    // Call object methods.
    reply, err := obj.Call("org.freedesktop.Notifications", "Notify",
        "dbus-tutorial", uint32(0), "",
        "dbus-tutorial", "You've been notified!",
	[]string{}, map[string]dbus.Variant{}, int32(-1))
    if err != nil {
        log.Fatal("Notification error:", err)
    }

    // Parse the reply message
    var notification_id uint32
    if err := reply.GetArgs(&notification_id); err != nil {
        log.Fatal(err)
    }
    log.Print("Notification id:", notification_id)
}
```
