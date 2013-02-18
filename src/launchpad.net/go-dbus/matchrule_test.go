package dbus

import . "launchpad.net/gocheck"

func (s *S) TestMatchRuleToString(c *C) {
	mr := MatchRule{
		Type:      TypeSignal,
		Interface: "org.freedesktop.DBus",
		Member:    "Foo",
		Path:      "/bar/foo"}
	c.Check(mr.String(), Equals, "type='signal',path='/bar/foo',interface='org.freedesktop.DBus',member='Foo'")

	// A rule that doesn't match the member
	mr = MatchRule{
		Type: TypeSignal,
		Interface: "com.example.Foo",
		Member: "Bar"}
	c.Check(mr.String(), Equals, "type='signal',interface='com.example.Foo',member='Bar'")
}

func (s *S) TestMatchRuleMatch(c *C) {
	msg := NewSignalMessage("", "org.freedesktop.DBus", "NameOwnerChanged")
	_ = msg.AppendArgs("com.example.Foo", "", ":2.0")

	mr := MatchRule{
		Type: TypeSignal,
		Interface: "org.freedesktop.DBus",
		Member: "NameOwnerChanged"}
	c.Check(mr._Match(msg), Equals, true)

	mr = MatchRule{
		Type: TypeSignal,
		Interface: "org.freedesktop.DBus",
		Member: "NameAcquired"}
	c.Check(mr._Match(msg), Equals, false)

	// Check matching against first argument.
	mr = MatchRule{
		Type: TypeSignal,
		Interface: "org.freedesktop.DBus",
		Member: "NameOwnerChanged",
		Arg0: "com.example.Foo"}
	c.Check(mr._Match(msg), Equals, true)
	mr.Arg0 = "com.example.Bar"
	c.Check(mr._Match(msg), Equals, false)
}
