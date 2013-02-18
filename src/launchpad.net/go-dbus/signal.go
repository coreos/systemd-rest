package dbus

import (
	"errors"
)

type SignalWatch struct {
	bus     *Connection
	rule    MatchRule
	handler func(*Message)

	// If the rule tries to match against a bus name as the
	// sender, we need to track the current owner of that name.
	nameWatch *NameWatch

	cancelled bool
}

// A structure to store the set of signal watches, keyed by object
// path, interface and member.
type signalWatchSet map[ObjectPath] map[string] map[string] []*SignalWatch

func (self signalWatchSet) Add(watch *SignalWatch) {
	byInterface, ok := self[watch.rule.Path]
	if !ok {
		byInterface = make(map[string] map[string] []*SignalWatch)
		self[watch.rule.Path] = byInterface
	}
	byMember, ok := byInterface[watch.rule.Interface]
	if !ok {
		byMember = make(map[string] []*SignalWatch)
		byInterface[watch.rule.Interface] = byMember
	}
	watches, ok := byMember[watch.rule.Member]
	if !ok {
		watches = make([]*SignalWatch, 0, 1)
	}
	byMember[watch.rule.Member] = append(watches, watch)
}

func (self signalWatchSet) Remove(watch *SignalWatch) bool {
	byInterface, ok := self[watch.rule.Path]
	if !ok {
		return false
	}
	byMember, ok := byInterface[watch.rule.Interface]
	if !ok {
		return false
	}
	watches, ok := byMember[watch.rule.Member]
	if !ok {
		return false
	}
	for i, other := range watches {
		if other == watch {
			// Truncate the watch slice, moving the item
			// at the end to the new location.
			watches[i] = watches[len(watches)-1]
			byMember[watch.rule.Member] = watches[:len(watches)-1]
			return true
		}
	}
	return false
}

func (self signalWatchSet) FindMatches(msg *Message) (matches []*SignalWatch) {
	pathKeys := []ObjectPath{""}
	if msg.Path != ObjectPath("") {
		pathKeys = append(pathKeys, msg.Path)
	}
	ifaceKeys := []string{""}
	if msg.Iface != "" {
		ifaceKeys = append(ifaceKeys, msg.Iface)
	}
	memberKeys := []string{""}
	if msg.Member != "" {
		memberKeys = append(memberKeys, msg.Member)
	}
	for _, path := range pathKeys {
		byInterface, ok := self[path]
		if !ok {
			continue
		}
		for _, iface := range ifaceKeys {
			byMember, ok := byInterface[iface]
			if !ok {
				continue
			}
			for _, member := range memberKeys {
				watches, ok := byMember[member]
				if !ok {
					continue
				}
				for _, watch := range watches {
					if watch.rule._Match(msg) {
						matches = append(matches, watch)
					}
				}
			}
		}
	}
	return
}

// Handle received signals.
func (p *Connection) WatchSignal(rule *MatchRule, handler func(*Message)) (*SignalWatch, error) {
	if rule.Type != TypeSignal {
		return nil, errors.New("Match rule is not for signals")
	}
	watch := &SignalWatch{bus: p, rule: *rule, handler: handler}

	// Does the rule match a bus name other than the daemon?
	if rule.Sender != "" && rule.Sender != BUS_DAEMON_NAME {
		var nameHandler func(string)
		if rule.Sender[0] == ':' {
			// For unique names, cancel the signal watch
			// when the name is lost.
			nameHandler = func (newOwner string) {
				if newOwner == "" {
					watch.Cancel()
				}
			}
		} else {
			// Otherwise, update the sender owner.
			nameHandler = func (newOwner string) {
				watch.rule.senderNameOwner = newOwner
			}
		}
		nameWatch, err := p.WatchName(rule.Sender, nameHandler)
		if err != nil {
			return nil, err
		}
		watch.nameWatch = nameWatch
	}
	if err := p.busProxy.AddMatch(rule.String()); err != nil {
		watch.nameWatch.Cancel()
		return nil, err
	}

	p.handlerMutex.Lock()
	p.signalMatchRules.Add(watch)
	p.handlerMutex.Unlock()
	return watch, nil
}

func (watch *SignalWatch) Cancel() error {
	if watch.cancelled {
		return nil
	}
	watch.cancelled = true
	watch.bus.handlerMutex.Lock()
	foundMatch := watch.bus.signalMatchRules.Remove(watch)
	watch.bus.handlerMutex.Unlock()

	if foundMatch {
		if err := watch.bus.busProxy.RemoveMatch(watch.rule.String()); err != nil {
			return err
		}
		if watch.nameWatch != nil {
			if err := watch.nameWatch.Cancel(); err != nil {
				return err
			}
		}
	}
	return nil
}
