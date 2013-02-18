package dbus

import "fmt"
import "strings"

// Matches all messages with equal type, interface, member, or path.
// Any missing/invalid fields are not matched against.
type MatchRule struct {
	Type      MessageType
	Sender    string
	Path      ObjectPath
	Interface string
	Member    string
	Arg0      string

	senderNameOwner string
}

// A string representation af the MatchRule (D-Bus variant map).
func (p *MatchRule) String() string {
	params := make([]string, 0, 6)
	if p.Type != TypeInvalid {
		params = append(params, fmt.Sprintf("type='%s'", p.Type))
	}
	if p.Sender != "" {
		params = append(params, fmt.Sprintf("sender='%s'", p.Sender))
	}
	if p.Path != "" {
		params = append(params, fmt.Sprintf("path='%s'", p.Path))
	}
	if p.Interface != "" {
		params = append(params, fmt.Sprintf("interface='%s'", p.Interface))
	}
	if p.Member != "" {
		params = append(params, fmt.Sprintf("member='%s'", p.Member))
	}
	if p.Arg0 != "" {
		params = append(params, fmt.Sprintf("arg0='%s'", p.Arg0))
	}
	return strings.Join(params, ",")
}

func (p *MatchRule) _Match(msg *Message) bool {
	if p.Type != TypeInvalid && p.Type != msg.Type {
		return false
	}
	if p.Sender != "" {
		if !(p.Sender == msg.Sender || p.senderNameOwner == msg.Sender) {
			return false
		}
	}
	if p.Path != "" && p.Path != msg.Path {
		return false
	}
	if p.Interface != "" && p.Interface != msg.Iface {
		return false
	}
	if p.Member != "" && p.Member != msg.Member {
		return false
	}
	if p.Arg0 != "" {
		var arg0 string
		if err := msg.GetArgs(&arg0); err != nil || arg0 != p.Arg0 {
			return false
		}
	}
	return true
}
