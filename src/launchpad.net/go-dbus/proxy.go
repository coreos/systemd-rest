package dbus

// This is not yet finished: it is an idea for what statically generated object bindings could look like.

type Introspectable struct {
	*ObjectProxy
}

func (o *Introspectable) Introspect() (data string, err error) {
	reply, err := o.Call("org.freedesktop.DBus.Introspectable", "Introspect")
	if err != nil {
		return
	}
	err = reply.GetArgs(&data)
	return
}

type Properties struct {
	*ObjectProxy
}

func (o *Properties) Get(interfaceName string, propertyName string) (value interface{}, err error) {
	reply, err := o.Call("org.freedesktop.DBus.Properties", "Get", interfaceName, propertyName)
	if err != nil {
		return
	}
	var variant Variant
	err = reply.GetArgs(&variant)
	value = variant.Value
	return
}

func (o *Properties) Set(interfaceName string, propertyName string, value interface{}) (err error) {
	_, err = o.Call("org.freedesktop.DBus.Properties", "Set", interfaceName, propertyName, Variant{value})
	return
}

func (o *Properties) GetAll(interfaceName string) (props map[string]Variant, err error) {
	reply, err := o.Call("org.freedesktop.DBus.Properties", "GetAll", interfaceName)
	if err != nil {
		return
	}
	err = reply.GetArgs(&props)
	return
}

type BusDaemon struct {
	*ObjectProxy
}

func (o *BusDaemon) Hello() (uniqueName string, err error) {
	reply, err := o.Call(BUS_DAEMON_IFACE, "Hello")
	if err != nil {
		return
	}
	err = reply.GetArgs(&uniqueName)
	return
}

func (o *BusDaemon) RequestName(name string, flags uint32) (result uint32, err error) {
	reply, err := o.Call(BUS_DAEMON_IFACE, "RequestName", name, flags)
	if err != nil {
		return
	}
	err = reply.GetArgs(&result)
	return
}

func (o *BusDaemon) ReleaseName(name string) (result uint32, err error) {
	reply, err := o.Call(BUS_DAEMON_IFACE, "ReleaseName", name)
	if err != nil {
		return
	}
	err = reply.GetArgs(&result)
	return
}

func (o *BusDaemon) ListQueuedOwners(name string) (owners []string, err error) {
	reply, err := o.Call(BUS_DAEMON_IFACE, "ListQueuedOwners", name)
	if err != nil {
		return
	}
	err = reply.GetArgs(&owners)
	return
}

func (o *BusDaemon) ListNames() (names []string, err error) {
	reply, err := o.Call(BUS_DAEMON_IFACE, "ListNames")
	if err != nil {
		return
	}
	err = reply.GetArgs(&names)
	return
}

func (o *BusDaemon) ListActivatableNames() (names []string, err error) {
	reply, err := o.Call(BUS_DAEMON_IFACE, "ListActivatableNames")
	if err != nil {
		return
	}
	err = reply.GetArgs(&names)
	return
}

func (o *BusDaemon) NameHasOwner(name string) (hasOwner bool, err error) {
	reply, err := o.Call(BUS_DAEMON_IFACE, "NameHasOwner", name)
	if err != nil {
		return
	}
	err = reply.GetArgs(&hasOwner)
	return
}

func (o *BusDaemon) StartServiceByName(name string, flags uint32) (result uint32, err error) {
	reply, err := o.Call(BUS_DAEMON_IFACE, "StartServiceByName", name, flags)
	if err != nil {
		return
	}
	err = reply.GetArgs(&result)
	return
}

func (o *BusDaemon) UpdateActivationEnvironment(env map[string]string) (err error) {
	_, err = o.Call(BUS_DAEMON_IFACE, "UpdateActivationEnvironment", env)
	return
}

func (o *BusDaemon) GetNameOwner(name string) (owner string, err error) {
	reply, err := o.Call(BUS_DAEMON_IFACE, "GetNameOwner", name)
	if err != nil {
		return
	}
	err = reply.GetArgs(&owner)
	return
}

func (o *BusDaemon) GetConnectionUnixUser(busName string) (user uint32, err error) {
	reply, err := o.Call(BUS_DAEMON_IFACE, "GetConnectionUnixUser", busName)
	if err != nil {
		return
	}
	err = reply.GetArgs(&user)
	return
}

func (o *BusDaemon) GetConnectionUnixProcessID(busName string) (process uint32, err error) {
	reply, err := o.Call(BUS_DAEMON_IFACE, "GetConnectionUnixProcessID", busName)
	if err != nil {
		return
	}
	err = reply.GetArgs(&process)
	return
}

func (o *BusDaemon) AddMatch(rule string) (err error) {
	_, err = o.Call(BUS_DAEMON_IFACE, "AddMatch", rule)
	return
}

func (o *BusDaemon) RemoveMatch(rule string) (err error) {
	_, err = o.Call(BUS_DAEMON_IFACE, "RemoveMatch", rule)
	return
}

func (o *BusDaemon) GetId() (busId string, err error) {
	reply, err := o.Call(BUS_DAEMON_IFACE, "GetId")
	if err != nil {
		return
	}
	err = reply.GetArgs(&busId)
	return
}
