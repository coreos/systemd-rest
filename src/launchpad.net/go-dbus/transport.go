package dbus

import (
	"errors"
	"net"
	"net/url"
	"strings"
)


type transport interface {
	Dial() (net.Conn, error)
}

func newTransport(address string) (transport, error) {
	if len(address) == 0 {
		return nil, errors.New("Unknown address type")
	}
	// Split the address into transport type and options.
	transportType := address[:strings.Index(address, ":")]
	options := make(map[string]string)
	for _, option := range strings.Split(address[len(transportType) + 1:], ",") {
		pair := strings.SplitN(option, "=", 2)
		key, err := url.QueryUnescape(pair[0])
		if err != nil {
			return nil, err
		}
		value, err := url.QueryUnescape(pair[1])
		if err != nil {
			return nil, err
		}
		options[key] = value
	}

	switch transportType {
	case "unix":
		if abstract, ok := options["abstract"]; ok {
			return &unixTransport{"@" + abstract}, nil
		} else if path, ok := options["path"]; ok {
			return &unixTransport{path}, nil
		} else {
			return nil, errors.New("unix transport requires 'path' or 'abstract' options")
		}
	case "tcp":
		address := options["host"] + ":" + options["port"]
		var family string
		switch options["family"] {
		case "", "ipv4":
			family = "tcp4"
		case "ipv6":
			family = "tcp6"
		default:
			return nil, errors.New("Unknown family for tcp transport: " + options["family"])
		}
		return &tcpTransport{address, family}, nil
	// These can be implemented later as needed
	case "nonce-tcp":
		// Like above, but with noncefile
	case "launchd":
		// Perform newTransport() on contents of
		// options["env"] environment variable
	case "systemd":
		// Socket Activation via LISTEN_PID/LISTEN_FDS
	case "unixexec":
		// exec a process with a socket hooked to stdin/stdout
	}

	return nil, errors.New("Unhandled transport type " + transportType)
}

type unixTransport struct {
	Address string
}

func (trans *unixTransport) Dial() (net.Conn, error) {
	return net.Dial("unix", trans.Address)
}

type tcpTransport struct {
	Address, Family string
}

func (trans *tcpTransport) Dial() (net.Conn, error) {
	return net.Dial(trans.Family, trans.Address)
}

