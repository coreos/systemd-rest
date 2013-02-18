package dbus

import . "launchpad.net/gocheck"

func (s *S) TestSignatureNextType(c *C) {
	// NextType() works for basic types
	for _, sig := range []Signature{"y", "b", "n", "q", "i", "u", "x", "t", "d", "s", "o", "g", "v", "h"} {
		next, err := sig.NextType(0)
		c.Check(next, Equals, 1)
		c.Check(err, Equals, nil)
	}

	// Unknown type code gives error
	next, err := Signature("_").NextType(0)
	c.Check(err, Not(Equals), nil)

	// Offset inside signature
	next, err = Signature("ii").NextType(1)
	c.Check(next, Equals, 2)
	c.Check(err, Equals, nil)

	// Error if there is no more type codes in signature
	next, err = Signature("i").NextType(1)
	c.Check(err, Not(Equals), nil)

	// Arrays consume their element type code
	next, err = Signature("ai").NextType(0)
	c.Check(next, Equals, 2)
	c.Check(err, Equals, nil)

	// Array without element type code gives error
	next, err = Signature("a").NextType(0)
	c.Check(err, Not(Equals), nil)

	// Structs are consumed entirely
	next, err = Signature("(isv)").NextType(0)
	c.Check(next, Equals, 5)
	c.Check(err, Equals, nil)

	// Incomplete struct gives error
	next, err = Signature("(isv").NextType(0)
	c.Check(err, Not(Equals), nil)

	// Dict entries have two contained type codes
	next, err = Signature("{ii}").NextType(0)
	c.Check(next, Equals, 4)
	c.Check(err, Equals, nil)

	next, err = Signature("{}").NextType(0)
	c.Check(err, Not(Equals), nil)
	next, err = Signature("{i}").NextType(0)
	c.Check(err, Not(Equals), nil)
	next, err = Signature("{iii}").NextType(0)
	c.Check(err, Not(Equals), nil)
	next, err = Signature("{ii").NextType(0)
	c.Check(err, Not(Equals), nil)

	// Now a recursive type combining the above.
	next, err = Signature("a{s(saax)}").NextType(0)
	c.Check(next, Equals, 10)
	c.Check(err, Equals, nil)
}

func (s *S) TestSignatureValidate(c *C) {
	c.Check(Signature("a{s(sax)}aav").Validate(), Equals, nil)
	c.Check(Signature("a").Validate(), Not(Equals), nil)
	c.Check(Signature("a(ii").Validate(), Not(Equals), nil)
}
