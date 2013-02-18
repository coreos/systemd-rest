package dbus

import (
	"encoding/binary"
	. "launchpad.net/gocheck"
)

func (s *S) TestEncoderAlign(c *C) {
	enc := newEncoder("", nil, binary.LittleEndian)
	enc.data.WriteByte(1)
	enc.align(1)
	c.Check(enc.data.Bytes(), DeepEquals, []byte{1})
	enc.align(2)
	c.Check(enc.data.Bytes(), DeepEquals, []byte{1, 0})
	enc.align(4)
	c.Check(enc.data.Bytes(), DeepEquals, []byte{1, 0, 0, 0})
	enc.align(8)
	c.Check(enc.data.Bytes(), DeepEquals, []byte{1, 0, 0, 0, 0, 0, 0, 0})
}

func (s *S) TestEncoderAppendByte(c *C) {
	enc := newEncoder("", nil, binary.LittleEndian)
	if err := enc.Append(byte(42)); err != nil {
		c.Error(err)
	}
	c.Check(enc.signature, Equals, Signature("y"))
	c.Check(enc.data.Bytes(), DeepEquals, []byte{42})
}

func (s *S) TestEncoderAppendBoolean(c *C) {
	enc := newEncoder("", nil, binary.LittleEndian)
	if err := enc.Append(true); err != nil {
		c.Error(err)
	}
	c.Check(enc.signature, Equals, Signature("b"))
	c.Check(enc.data.Bytes(), DeepEquals, []byte{1, 0, 0, 0})
}

func (s *S) TestEncoderAppendInt16(c *C) {
	enc := newEncoder("", nil, binary.LittleEndian)
	if err := enc.Append(int16(42)); err != nil {
		c.Error(err)
	}
	c.Check(enc.signature, Equals, Signature("n"))
	c.Check(enc.data.Bytes(), DeepEquals, []byte{42, 0})
}

func (s *S) TestEncoderAppendUint16(c *C) {
	enc := newEncoder("", nil, binary.LittleEndian)
	if err := enc.Append(uint16(42)); err != nil {
		c.Error(err)
	}
	c.Check(enc.signature, Equals, Signature("q"))
	c.Check(enc.data.Bytes(), DeepEquals, []byte{42, 0})
}

func (s *S) TestEncoderAppendInt32(c *C) {
	enc := newEncoder("", nil, binary.LittleEndian)
	if err := enc.Append(int32(42)); err != nil {
		c.Error(err)
	}
	c.Check(enc.signature, Equals, Signature("i"))
	c.Check(enc.data.Bytes(), DeepEquals, []byte{42, 0, 0, 0})
}

func (s *S) TestEncoderAppendUint32(c *C) {
	enc := newEncoder("", nil, binary.LittleEndian)
	if err := enc.Append(uint32(42)); err != nil {
		c.Error(err)
	}
	c.Check(enc.signature, Equals, Signature("u"))
	c.Check(enc.data.Bytes(), DeepEquals, []byte{42, 0, 0, 0})
}

func (s *S) TestEncoderAppendInt64(c *C) {
	enc := newEncoder("", nil, binary.LittleEndian)
	if err := enc.Append(int64(42)); err != nil {
		c.Error(err)
	}
	c.Check(enc.signature, Equals, Signature("x"))
	c.Check(enc.data.Bytes(), DeepEquals, []byte{42, 0, 0, 0, 0, 0, 0, 0})
}

func (s *S) TestEncoderAppendUint64(c *C) {
	enc := newEncoder("", nil, binary.LittleEndian)
	if err := enc.Append(uint64(42)); err != nil {
		c.Error(err)
	}
	c.Check(enc.signature, Equals, Signature("t"))
	c.Check(enc.data.Bytes(), DeepEquals, []byte{42, 0, 0, 0, 0, 0, 0, 0})
}

func (s *S) TestEncoderAppendFloat64(c *C) {
	enc := newEncoder("", nil, binary.LittleEndian)
	if err := enc.Append(float64(42.0)); err != nil {
		c.Error(err)
	}
	c.Check(enc.signature, Equals, Signature("d"))
	c.Check(enc.data.Bytes(), DeepEquals, []byte{0, 0, 0, 0, 0, 0, 69, 64})
}

func (s *S) TestEncoderAppendString(c *C) {
	enc := newEncoder("", nil, binary.LittleEndian)
	if err := enc.Append("hello"); err != nil {
		c.Error(err)
	}
	c.Check(enc.signature, Equals, Signature("s"))
	c.Check(enc.data.Bytes(), DeepEquals, []byte{
		5, 0, 0, 0,              // Length
		'h', 'e', 'l', 'l', 'o', // "hello"
		0})                      // nul termination
}

func (s *S) TestEncoderAppendObjectPath(c *C) {
	enc := newEncoder("", nil, binary.LittleEndian)
	if err := enc.Append(ObjectPath("/foo")); err != nil {
		c.Error(err)
	}
	c.Check(enc.signature, Equals, Signature("o"))
	c.Check(enc.data.Bytes(), DeepEquals, []byte{
		4, 0, 0, 0,         // Length
		'/', 'f', 'o', 'o', // ObjectPath("/foo")
		0})                 // nul termination
}

type testObject struct {}
func (f *testObject) GetObjectPath() ObjectPath {
	return ObjectPath("/foo")
}

func (s *S) TestEncoderAppendObject(c *C) {
	enc := newEncoder("", nil, binary.LittleEndian)
	if err := enc.Append(&testObject{}); err != nil {
		c.Error(err)
	}
	c.Check(enc.signature, Equals, Signature("o"))
	c.Check(enc.data.Bytes(), DeepEquals, []byte{
		4, 0, 0, 0,         // Length
		'/', 'f', 'o', 'o', // ObjectPath("/foo")
		0})                 // nul termination
}

func (s *S) TestEncoderAppendSignature(c *C) {
	enc := newEncoder("", nil, binary.LittleEndian)
	if err := enc.Append(Signature("a{si}")); err != nil {
		c.Error(err)
	}
	c.Check(enc.signature, Equals, Signature("g"))
	c.Check(enc.data.Bytes(), DeepEquals, []byte{
		5,                       // Length
		'a', '{', 's', 'i', '}', // Signature("a{si}")
		0})                      // nul termination
}

func (s *S) TestEncoderAppendArray(c *C) {
	enc := newEncoder("", nil, binary.LittleEndian)
	if err := enc.Append([]int32{42, 420}); err != nil {
		c.Error(err)
	}
	c.Check(enc.signature, Equals, Signature("ai"))
	c.Check(enc.data.Bytes(), DeepEquals, []byte{
		8, 0, 0, 0,    // Length
		42, 0, 0, 0,   // int32(42)
		164, 1, 0, 0}) // int32(420)
}

func (s *S) TestEncoderAppendArrayLengthAlignment(c *C) {
	enc := newEncoder("", nil, binary.LittleEndian)
	// append a byte, which means we are no longer aligned.
	c.Assert(enc.Append(byte(1)), Equals, nil)
	// Now create an array.
	c.Check(enc.Append([]uint32{42}), Equals, nil)
	c.Check(enc.signature, Equals, Signature("yau"))
	c.Check(enc.data.Bytes(), DeepEquals, []byte{
		1,            // byte(1)
		0, 0, 0,      // padding
		4, 0, 0, 0,   // array length
		42, 0, 0, 0}) // uint32(42)
}

func (s *S) TestEncoderAppendArrayPaddingAfterLength(c *C) {
	enc := newEncoder("", nil, binary.LittleEndian)
	// Now create an array with alignment 8 values.
	c.Check(enc.Append([]int64{42}), Equals, nil)
	c.Check(enc.signature, Equals, Signature("ax"))
	c.Check(enc.data.Bytes(), DeepEquals, []byte{
		8, 0, 0, 0,   // array length (not including padding)
                0, 0, 0, 0,   // padding
		42, 0, 0, 0, 0, 0, 0, 0}) // int64(42)

	// The padding is needed, even if there are no elements in the array.
	enc = newEncoder("", nil, binary.LittleEndian)
	c.Check(enc.Append([]int64{}), Equals, nil)
	c.Check(enc.signature, Equals, Signature("ax"))
	c.Check(enc.data.Bytes(), DeepEquals, []byte{
		0, 0, 0, 0,   // array length (not including padding)
                0, 0, 0, 0})  // padding
}

func (s *S) TestEncoderAppendMap(c *C) {
	enc := newEncoder("", nil, binary.LittleEndian)
	if err := enc.Append(map[string]bool{"true": true}); err != nil {
		c.Error(err)
	}
	c.Check(enc.signature, Equals, Signature("a{sb}"))
	c.Check(enc.data.Bytes(), DeepEquals, []byte{
		16, 0, 0, 0,                       // array content length
		0, 0, 0, 0,                        // padding to 8 bytes
		4, 0, 0, 0, 't', 'r', 'u', 'e', 0, // "true"
		0, 0, 0,                           // padding to 4 bytes
		1, 0, 0, 0})                       // true
}

func (s *S) TestEncoderAppendMapAlignment(c *C) {
	enc := newEncoder("", nil, binary.LittleEndian)
	// append a byte, which means we are no longer aligned.
	c.Assert(enc.Append(byte(1)), Equals, nil)

	c.Check(enc.Append(map[string]bool{"true": true}), Equals, nil)
	c.Check(enc.signature, Equals, Signature("ya{sb}"))
	c.Check(enc.data.Bytes(), DeepEquals, []byte{
		1,                                 // byte(1)
		0, 0, 0,                           // padding
		16, 0, 0, 0,                       // array content length
		4, 0, 0, 0, 't', 'r', 'u', 'e', 0, // "true"
		0, 0, 0,                           // padding to 4 bytes
		1, 0, 0, 0})                       // true
}

func (s *S) TestEncoderAppendStruct(c *C) {
	enc := newEncoder("", nil, binary.LittleEndian)
	type sample struct {
		one int32
		two string
	}
	if err := enc.Append(&sample{42, "hello"}); err != nil {
		c.Error(err)
	}
	c.Check(enc.signature, Equals, Signature("(is)"))
	c.Check(enc.data.Bytes(), DeepEquals, []byte{
		42, 0, 0, 0,
		5, 0, 0, 0, 'h', 'e' , 'l', 'l', 'o', 0})
}

func (s *S) TestEncoderAppendVariant(c *C) {
	enc := newEncoder("", nil, binary.LittleEndian)
	if err := enc.Append(&Variant{int32(42)}); err != nil {
		c.Error(err)
	}
	c.Check(enc.signature, Equals, Signature("v"))
	c.Check(enc.data.Bytes(), DeepEquals, []byte{
		1, 'i', 0,    // Signature("i")
		0,            // padding to 4 bytes
		42, 0, 0, 0}) // int32(42)
}

func (s *S) TestEncoderAppendAlignment(c *C) {
	enc := newEncoder("", nil, binary.LittleEndian)
	if err := enc.Append(byte(42), int16(42), true, int32(42), int64(42)); err != nil {
		c.Error(err)
	}
	c.Check(enc.signature, Equals, Signature("ynbix"))
	c.Check(enc.data.Bytes(), DeepEquals, []byte{
		42,                       // byte(42)
		0,                        // padding to 2 bytes
		42, 0,                    // int16(42)
		1, 0, 0, 0,               // true
		42, 0, 0, 0,              // int32(42)
		0, 0, 0, 0,               // padding to 8 bytes
		42, 0, 0, 0, 0, 0, 0, 0}) // int64(42)
}
