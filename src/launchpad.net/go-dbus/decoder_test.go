package dbus

import "encoding/binary"
import . "launchpad.net/gocheck"

func (s *S) TestDecoderDecodeByte(c *C) {
	dec := newDecoder("yy", []byte{42, 100}, binary.LittleEndian)
	var value1 byte
	var value2 interface{}
	if err := dec.Decode(&value1, &value2); err != nil {
		c.Error(err)
	}
	c.Check(value1, Equals, byte(42))
	c.Check(value2, Equals, byte(100))
	c.Check(dec.dataOffset, Equals, 2)
	c.Check(dec.sigOffset, Equals, 2)
}

func (s *S) TestDecoderDecodeBool(c *C) {
	dec := newDecoder("bb", []byte{0, 0, 0, 0, 1, 0, 0, 0}, binary.LittleEndian)
	var value1 bool
	var value2 interface{}
	if err := dec.Decode(&value1, &value2); err != nil {
		c.Error(err)
	}
	c.Check(value1, Equals, false)
	c.Check(value2, Equals, true)
	c.Check(dec.dataOffset, Equals, 8)
	c.Check(dec.sigOffset, Equals, 2)
}

func (s *S) TestDecoderDecodeInt16(c *C) {
	dec := newDecoder("nn", []byte{42, 0, 100, 0}, binary.LittleEndian)
	var value1 int16
	var value2 interface{}
	if err := dec.Decode(&value1, &value2); err != nil {
		c.Error(err)
	}
	c.Check(value1, Equals, int16(42))
	c.Check(value2, Equals, int16(100))
	c.Check(dec.dataOffset, Equals, 4)
	c.Check(dec.sigOffset, Equals, 2)
}

func (s *S) TestDecoderDecodeUint16(c *C) {
	dec := newDecoder("qq", []byte{42, 0, 100, 0}, binary.LittleEndian)
	var value1 uint16
	var value2 interface{}
	if err := dec.Decode(&value1, &value2); err != nil {
		c.Error(err)
	}
	c.Check(value1, Equals, uint16(42))
	c.Check(value2, Equals, uint16(100))
	c.Check(dec.dataOffset, Equals, 4)
	c.Check(dec.sigOffset, Equals, 2)
}

func (s *S) TestDecoderDecodeInt32(c *C) {
	dec := newDecoder("ii", []byte{42, 0, 0, 0, 100, 0, 0, 0}, binary.LittleEndian)
	var value1 int32
	var value2 interface{}
	if err := dec.Decode(&value1, &value2); err != nil {
		c.Error(err)
	}
	c.Check(value1, Equals, int32(42))
	c.Check(value2, Equals, int32(100))
	c.Check(dec.dataOffset, Equals, 8)
	c.Check(dec.sigOffset, Equals, 2)
}

func (s *S) TestDecoderDecodeUint32(c *C) {
	dec := newDecoder("uu", []byte{42, 0, 0, 0, 100, 0, 0, 0}, binary.LittleEndian)
	var value1 uint32
	var value2 interface{}
	if err := dec.Decode(&value1, &value2); err != nil {
		c.Error(err)
	}
	c.Check(value1, Equals, uint32(42))
	c.Check(value2, Equals, uint32(100))
	c.Check(dec.dataOffset, Equals, 8)
	c.Check(dec.sigOffset, Equals, 2)
}

func (s *S) TestDecoderDecodeInt64(c *C) {
	dec := newDecoder("xx", []byte{42, 0, 0, 0, 0, 0, 0, 0, 100, 0, 0, 0, 0, 0, 0, 0}, binary.LittleEndian)
	var value1 int64
	var value2 interface{}
	if err := dec.Decode(&value1, &value2); err != nil {
		c.Error(err)
	}
	c.Check(value1, Equals, int64(42))
	c.Check(value2, Equals, int64(100))
	c.Check(dec.dataOffset, Equals, 16)
	c.Check(dec.sigOffset, Equals, 2)
}

func (s *S) TestDecoderDecodeUint64(c *C) {
	dec := newDecoder("tt", []byte{42, 0, 0, 0, 0, 0, 0, 0, 100, 0, 0, 0, 0, 0, 0, 0}, binary.LittleEndian)
	var value1 uint64
	var value2 interface{}
	if err := dec.Decode(&value1, &value2); err != nil {
		c.Error(err)
	}
	c.Check(value1, Equals, uint64(42))
	c.Check(value2, Equals, uint64(100))
	c.Check(dec.dataOffset, Equals, 16)
	c.Check(dec.sigOffset, Equals, 2)
}

func (s *S) TestDecoderDecodeFloat64(c *C) {
	dec := newDecoder("dd", []byte{0, 0, 0, 0, 0, 0, 69, 64, 0, 0, 0, 0, 0, 0, 89, 64}, binary.LittleEndian)
	var value1 float64
	var value2 interface{}
	if err := dec.Decode(&value1, &value2); err != nil {
		c.Error(err)
	}
	c.Check(value1, Equals, float64(42))
	c.Check(value2, Equals, float64(100))
	c.Check(dec.dataOffset, Equals, 16)
	c.Check(dec.sigOffset, Equals, 2)
}

func (s *S) TestDecoderDecodeString(c *C) {
	dec := newDecoder("ss", []byte{
		5, 0, 0, 0,                  // len("hello")
		'h', 'e', 'l', 'l', 'o', 0,  // "hello"
		0, 0,                        // padding
		5, 0, 0, 0,                  // len("world")
		'w', 'o', 'r', 'l', 'd', 0}, // "world"
		binary.LittleEndian)
	var value1 string
	var value2 interface{}
	if err := dec.Decode(&value1, &value2); err != nil {
		c.Error(err)
	}
	c.Check(value1, Equals, "hello")
	c.Check(value2, Equals, "world")
	c.Check(dec.dataOffset, Equals, 22)
	c.Check(dec.sigOffset, Equals, 2)
}

func (s *S) TestDecoderDecodeObjectPath(c *C) {
	dec := newDecoder("oo", []byte{
		4, 0, 0, 0,             // len("/foo")
		'/', 'f', 'o', 'o', 0,  // ObjectPath("/foo")
		0, 0, 0,                // padding
		4, 0, 0, 0,             // len("/bar")
		'/', 'b', 'a', 'r', 0}, // ObjectPath("/bar")
		binary.LittleEndian)
	var value1 ObjectPath
	var value2 interface{}
	if err := dec.Decode(&value1, &value2); err != nil {
		c.Error(err)
	}
	c.Check(value1, Equals, ObjectPath("/foo"))
	c.Check(value2, Equals, ObjectPath("/bar"))
	c.Check(dec.dataOffset, Equals, 21)
	c.Check(dec.sigOffset, Equals, 2)
}

func (s *S) TestDecoderDecodeSignature(c *C) {
	dec := newDecoder("gg", []byte{
		8,                                         // len("a{s(iv)}")
		'a', '{', 's', '(', 'i', 'v', ')', '}', 0, // Signature("a{s(iv)}")
		4,                                         // len("asvi")
		'a', 's', 'v', 'i', 0},                    // Signature("asvi")
		binary.LittleEndian)
	var value1 Signature
	var value2 interface{}
	if err := dec.Decode(&value1, &value2); err != nil {
		c.Error(err)
	}
	c.Check(value1, Equals, Signature("a{s(iv)}"))
	c.Check(value2, Equals, Signature("asvi"))
	c.Check(dec.dataOffset, Equals, 16)
	c.Check(dec.sigOffset, Equals, 2)
}

func (s *S) TestDecoderDecodeArray(c *C) {
	dec := newDecoder("ai", []byte{
		8, 0, 0, 0,    // array length
		42, 0, 0, 0,   // int32(42)
		100, 0, 0, 0}, // int32(100)
		binary.LittleEndian)
	// Decode as an array
	var value1 [2]int32
	if err := dec.Decode(&value1); err != nil {
		c.Error("Decode as array:", err)
	}
	c.Check(dec.dataOffset, Equals, 12)
	c.Check(dec.sigOffset, Equals, 2)
	c.Check(value1[0], Equals, int32(42))
	c.Check(value1[1], Equals, int32(100))

	// Decode as a slice
	dec.dataOffset = 0
	dec.sigOffset = 0
	var value2 []int32
	if err := dec.Decode(&value2); err != nil {
		c.Error("Decode as slice:", err)
	}
	c.Check(value2, DeepEquals, []int32{42, 100})

	// Decode as blank interface
	dec.dataOffset = 0
	dec.sigOffset = 0
	var value3 interface{}
	if err := dec.Decode(&value3); err != nil {
		c.Error("Decode as interface:", err)
	}
	c.Check(value3, DeepEquals, []interface{}{int32(42), int32(100)})
}

func (s *S) TestDecoderDecodeEmptyArray(c *C) {
	dec := newDecoder("ai", []byte{
		0, 0, 0, 0}, // array length
		binary.LittleEndian)
	var value []int32
	c.Check(dec.Decode(&value), Equals, nil)
	c.Check(dec.dataOffset, Equals, 4)
	c.Check(dec.sigOffset, Equals, 2)
	c.Check(value, DeepEquals, []int32{})
}

func (s *S) TestDecoderDecodeArrayPaddingAfterLength(c *C) {
	dec := newDecoder("ax", []byte{
		8, 0, 0, 0,               // array length
                0, 0, 0, 0,               // padding
                42, 0, 0, 0, 0, 0, 0, 0}, // uint64(42)
		binary.LittleEndian)
	var value []int64
	c.Check(dec.Decode(&value), Equals, nil)
	c.Check(dec.dataOffset, Equals, 16)
	c.Check(dec.sigOffset, Equals, 2)
	c.Check(value, DeepEquals, []int64{42})

	// This padding exists even for empty arays
	dec = newDecoder("ax", []byte{
		0, 0, 0, 0,  // array length
                0, 0, 0, 0}, // padding
		binary.LittleEndian)
	c.Check(dec.Decode(&value), Equals, nil)
	c.Check(dec.dataOffset, Equals, 8)
	c.Check(dec.sigOffset, Equals, 2)
	c.Check(value, DeepEquals, []int64{})
}

func (s *S) TestDecoderDecodeMap(c *C) {
	dec := newDecoder("a{si}", []byte{
		36, 0, 0, 0,      // array length
		0, 0, 0, 0,       // padding
                3, 0, 0, 0,       // len("one")
                'o', 'n', 'e', 0, // "one"
                1, 0, 0, 0,       // int32(1)
                0, 0, 0, 0,       // padding
                9, 0, 0, 0,       // len("forty two")
                'f', 'o', 'r', 't', 'y', ' ', 't', 'w', 'o', 0,
                0, 0,             // padding
		42, 0, 0, 0},     // int32(42)
		binary.LittleEndian)
	var value map[string]int32
	c.Check(dec.Decode(&value), Equals, nil)
	c.Check(len(value), Equals, 2)
	c.Check(value["one"], Equals, int32(1))
	c.Check(value["forty two"], Equals, int32(42))
}

func (s *S) TestDecoderDecodeStruct(c *C) {
	dec := newDecoder("(si)", []byte{
		5, 0, 0, 0,                 // len("hello")
                'h', 'e', 'l', 'l', 'o', 0, // "hello"
		0, 0,                       // padding
                42, 0, 0, 0},               // int32(42)
		binary.LittleEndian)

	type Dummy struct {
		S string
		I int32
	}
	// Decode as structure
	var value1 Dummy
	if err := dec.Decode(&value1); err != nil {
		c.Error("Decode as structure:", err)
	}
	c.Check(dec.dataOffset, Equals, 16)
	c.Check(dec.sigOffset, Equals, 4)
	c.Check(value1, DeepEquals, Dummy{"hello", 42})

	// Decode as pointer to structure
	dec.dataOffset = 0
	dec.sigOffset = 0
	var value2 *Dummy
	if err := dec.Decode(&value2); err != nil {
		c.Error("Decode as structure pointer:", err)
	}
	c.Check(value2, DeepEquals, &Dummy{"hello", 42})

	// Decode as blank interface
	dec.dataOffset = 0
	dec.sigOffset = 0
	var value3 interface{}
	if err := dec.Decode(&value3); err != nil {
		c.Error("Decode as interface:", err)
	}
	c.Check(value3, DeepEquals, []interface{}{"hello", int32(42)})
}

func (s *S) TestDecoderDecodeVariant(c *C) {
	dec := newDecoder("v", []byte{
		1,            // len("i")
		'i', 0,       // Signature("i")
		0,            // padding
                42, 0, 0, 0}, // int32(42)
		binary.LittleEndian)

	var value1 Variant
	if err := dec.Decode(&value1); err != nil {
		c.Error("Decode as Variant:", err)
	}
	c.Check(dec.dataOffset, Equals, 8)
	c.Check(dec.sigOffset, Equals, 1)
	c.Check(value1, DeepEquals, Variant{int32(42)})

	// Decode as pointer to Variant
	dec.dataOffset = 0
	dec.sigOffset = 0
	var value2 *Variant
	if err := dec.Decode(&value2); err != nil {
		c.Error("Decode as *Variant:", err)
	}
	c.Check(value2, DeepEquals, &Variant{int32(42)})

	// Decode as pointer to blank interface
	dec.dataOffset = 0
	dec.sigOffset = 0
	var value3 interface{}
	if err := dec.Decode(&value3); err != nil {
		c.Error("Decode as interface:", err)
	}
	c.Check(value3, DeepEquals, &Variant{int32(42)})
}
