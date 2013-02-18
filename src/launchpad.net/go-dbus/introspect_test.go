package dbus

import (
	. "launchpad.net/gocheck"
)

var introStr = `
        <!DOCTYPE node PUBLIC "-//freedesktop//DTD D-BUS Object Introspection 1.0//EN"
         "http://www.freedesktop.org/standards/dbus/1.0/introspect.dtd">
        <node name="/org/freedesktop/sample_object">
          <interface name="org.freedesktop.SampleInterface">
            <method name="Frobate">
              <arg name="foo" type="i" direction="in"/>
              <arg name="bar" type="s" direction="out"/>
              <arg name="baz" type="a{us}" direction="out"/>
              <annotation name="org.freedesktop.DBus.Deprecated" value="true"/>
            </method>
            <method name="Bazify">
              <arg name="bar" type="(iiu)" direction="in"/>
              <arg name="bar" type="v" direction="out"/>
            </method>
            <method name="Mogrify">
              <arg name="bar" type="(iiav)" direction="in"/>
            </method>
            <signal name="Changed">
              <arg name="new_value" type="b"/>
            </signal>
            <property name="Bar" type="y" access="readwrite"/>
          </interface>
          <node name="child_of_sample_object"/>
          <node name="another_child_of_sample_object"/>
       </node>
`

func (s *S) TestIntrospect(c *C) {
	intro, err := NewIntrospect(introStr)
	c.Assert(err, Equals, nil)
	c.Assert(intro, Not(Equals), nil)

	intf := intro.GetInterfaceData("org.freedesktop.SampleInterface")
	c.Assert(intf,  Not(Equals), nil)
	c.Check(intf.GetName(), Equals, "org.freedesktop.SampleInterface")

	meth := intf.GetMethodData("Frobate")
	c.Assert(meth, Not(Equals), nil)
	c.Check(meth.GetOutSignature(), Equals, Signature("sa{us}"))

	nilmeth := intf.GetMethodData("Hoo") // unknown method name
	c.Check(nilmeth, Equals, nil)

	signal := intf.GetSignalData("Changed")
	c.Assert(signal, Not(Equals), nil)
	c.Check(signal.GetSignature(), Equals, Signature("b"))

	nilsignal := intf.GetSignalData("Hoo") // unknown signal name
	c.Check(nilsignal, Equals, nil)
}
