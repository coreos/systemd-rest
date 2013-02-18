package dbus

import (
	"encoding/binary"
	"errors"
	"math"
	"reflect"
)

type decoder struct {
	signature Signature
	data []byte
	order binary.ByteOrder

	dataOffset, sigOffset int
}

var (
	bufferOverrunError = errors.New("Buffer too small")
	signatureOverrunError = errors.New("Signature too small"))

func newDecoder(signature Signature, data []byte, order binary.ByteOrder) *decoder {
	return &decoder{signature: signature, data: data, order: order}
}

func (self *decoder) align(alignment int) {
	inc := -self.dataOffset % alignment
	if inc < 0 {
		inc += alignment
	}
	self.dataOffset += inc
}

func (self *decoder) Decode(args ...interface{}) error {
	for _, arg := range args {
		v := reflect.ValueOf(arg)
		// We expect to be given pointers here, so the caller
		// can see the decoded values.
		if v.Kind() != reflect.Ptr {
			return errors.New("arguments to Decode should be pointers")
		}
		if err := self.decodeValue(v.Elem()); err != nil {
			return err
		}
	}
	return nil
}

func (self *decoder) HasMore() bool {
	return self.sigOffset < len(self.signature)
}

func (self *decoder) Remainder() []byte {
	return self.data[self.dataOffset:]
}

func (self *decoder) readByte() (byte, error) {
	if len(self.data) < self.dataOffset + 1 {
		return 0, bufferOverrunError
	}
	value := self.data[self.dataOffset]
	self.dataOffset += 1
	return value, nil
}

func (self *decoder) readInt16() (int16, error) {
	self.align(2)
	if len(self.data) < self.dataOffset + 2 {
		return 0, bufferOverrunError
	}
	value := int16(self.order.Uint16(self.data[self.dataOffset:]))
	self.dataOffset += 2
	return value, nil
}

func (self *decoder) readUint16() (uint16, error) {
	self.align(2)
	if len(self.data) < self.dataOffset + 2 {
		return 0, bufferOverrunError
	}
	value := self.order.Uint16(self.data[self.dataOffset:])
	self.dataOffset += 2
	return value, nil
}

func (self *decoder) readInt32() (int32, error) {
	self.align(4)
	if len(self.data) < self.dataOffset + 4 {
		return 0, bufferOverrunError
	}
	value := int32(self.order.Uint32(self.data[self.dataOffset:]))
	self.dataOffset += 4
	return value, nil
}

func (self *decoder) readUint32() (uint32, error) {
	self.align(4)
	if len(self.data) < self.dataOffset + 4 {
		return 0, bufferOverrunError
	}
	value := self.order.Uint32(self.data[self.dataOffset:])
	self.dataOffset += 4
	return value, nil
}

func (self *decoder) readInt64() (int64, error) {
	self.align(8)
	if len(self.data) < self.dataOffset + 8 {
		return 0, bufferOverrunError
	}
		value := int64(self.order.Uint64(self.data[self.dataOffset:]))
	self.dataOffset += 8
	return value, nil
}

func (self *decoder) readUint64() (uint64, error) {
	self.align(8)
	if len(self.data) < self.dataOffset + 8 {
		return 0, bufferOverrunError
	}
	value := self.order.Uint64(self.data[self.dataOffset:])
	self.dataOffset += 8
	return value, nil
}

func (self *decoder) readFloat64() (float64, error) {
	value, err := self.readUint64()
	return math.Float64frombits(value), err
}

func (self *decoder) readString() (string, error) {
	length, err := self.readUint32()
	if err != nil {
		return "", err
	}
	// One extra byte for null termination
	if len(self.data) < self.dataOffset + int(length) + 1 {
		return "", bufferOverrunError
	}
	value := string(self.data[self.dataOffset:self.dataOffset + int(length)])
	self.dataOffset += int(length) + 1
	return value, nil
}

func (self *decoder) readSignature() (Signature, error) {
	length, err := self.readByte()
	if err != nil {
		return "", err
	}
	// One extra byte for null termination
	if len(self.data) < self.dataOffset + int(length) + 1 {
		return "", bufferOverrunError
	}
	value := Signature(self.data[self.dataOffset:self.dataOffset + int(length)])
	self.dataOffset += int(length) + 1
	return value, nil
}

func (self *decoder) decodeValue(v reflect.Value) error {
	if len(self.signature) < self.sigOffset {
		return signatureOverrunError
	}
	sigCode := self.signature[self.sigOffset]
	self.sigOffset += 1
	switch sigCode {
	case 'y':
		value, err := self.readByte()
		if err != nil {
			return err
		}
		switch {
		case v.Kind() == reflect.Uint8:
			v.SetUint(uint64(value))
			return nil
		case typeBlankInterface.AssignableTo(v.Type()):
			v.Set(reflect.ValueOf(value))
			return nil
		}
	case 'b':
		value, err := self.readUint32()
		if err != nil {
			return err
		}
		switch {
		case v.Kind() == reflect.Bool:
			v.SetBool(value != 0)
			return nil
		case typeBlankInterface.AssignableTo(v.Type()):
			v.Set(reflect.ValueOf(value != 0))
			return nil
		}
	case 'n':
		value, err := self.readInt16()
		if err != nil {
			return err
		}
		switch {
		case v.Kind() == reflect.Int16:
			v.SetInt(int64(value))
			return nil
		case typeBlankInterface.AssignableTo(v.Type()):
			v.Set(reflect.ValueOf(value))
			return nil
		}
	case 'q':
		value, err := self.readUint16()
		if err != nil {
			return err
		}
		switch {
		case v.Kind() == reflect.Uint16:
			v.SetUint(uint64(value))
			return nil
		case typeBlankInterface.AssignableTo(v.Type()):
			v.Set(reflect.ValueOf(value))
			return nil
		}
	case 'i':
		value, err := self.readInt32()
		if err != nil {
			return err
		}
		switch {
		case v.Kind() == reflect.Int32:
			v.SetInt(int64(value))
			return nil
		case typeBlankInterface.AssignableTo(v.Type()):
			v.Set(reflect.ValueOf(value))
			return nil
		}
	case 'u':
		value, err := self.readUint32()
		if err != nil {
			return err
		}
		switch {
		case v.Kind() == reflect.Uint32:
			v.SetUint(uint64(value))
			return nil
		case typeBlankInterface.AssignableTo(v.Type()):
			v.Set(reflect.ValueOf(value))
			return nil
		}
	case 'x':
		value, err := self.readInt64()
		if err != nil {
			return err
		}
		switch {
		case v.Kind() == reflect.Int64:
			v.SetInt(int64(value))
			return nil
		case typeBlankInterface.AssignableTo(v.Type()):
			v.Set(reflect.ValueOf(value))
			return nil
		}
	case 't':
		value, err := self.readUint64()
		if err != nil {
			return err
		}
		switch {
		case v.Kind() == reflect.Uint64:
			v.SetUint(uint64(value))
			return nil
		case typeBlankInterface.AssignableTo(v.Type()):
			v.Set(reflect.ValueOf(value))
			return nil
		}
	case 'd':
		value, err := self.readFloat64()
		if err != nil {
			return err
		}
		switch {
		case v.Kind() == reflect.Float64:
			v.SetFloat(value)
			return nil
		case typeBlankInterface.AssignableTo(v.Type()):
			v.Set(reflect.ValueOf(value))
			return nil
		}
	case 's':
		value, err := self.readString()
		if err != nil {
			return err
		}
		switch {
		case v.Kind() == reflect.String:
			v.SetString(value)
			return nil
		case typeBlankInterface.AssignableTo(v.Type()):
			v.Set(reflect.ValueOf(value))
			return nil
		}
	case 'o':
		value, err := self.readString()
		if err != nil {
			return err
		}
		switch {
		case v.Kind() == reflect.String:
			v.SetString(value)
			return nil
		case typeBlankInterface.AssignableTo(v.Type()):
			v.Set(reflect.ValueOf(ObjectPath(value)))
			return nil
		}
	case 'g':
		value, err := self.readSignature()
		if err != nil {
			return err
		}
		switch {
		case v.Kind() == reflect.String:
			v.SetString(string(value))
			return nil
		case typeBlankInterface.AssignableTo(v.Type()):
			v.Set(reflect.ValueOf(value))
			return nil
		}
	case 'a':
		length, err := self.readUint32()
		if err != nil {
			return err
		}
		elemSigOffset := self.sigOffset
		afterElemOffset, err := self.signature.NextType(elemSigOffset)
		if err != nil {
			return err
		}
		// Adjust data offset so we are aligned to read array
		// elements.  Anything with an alignment of 4 or less
		// will already be aligned due to reading the length.
		switch self.signature[self.sigOffset] {
		case 'x', 't', 'd', '(', '{':
			self.align(8)
		}
		arrayEnd := self.dataOffset + int(length)
		if len(self.data) < arrayEnd {
			return bufferOverrunError
		}
		switch {
		case v.Kind() == reflect.Array:
			for i := 0; self.dataOffset < arrayEnd; i++ {
				// Reset signature offset to the array element.
				self.sigOffset = elemSigOffset
				if err := self.decodeValue(v.Index(i)); err != nil {
					return err
				}
			}
			self.sigOffset = afterElemOffset
			return nil
		case v.Kind() == reflect.Slice:
			if v.IsNil() {
				v.Set(reflect.MakeSlice(v.Type(), 0, 0))
			}
			v.SetLen(0)
			for self.dataOffset < arrayEnd {
				// Reset signature offset to the array element.
				self.sigOffset = elemSigOffset
				elem := reflect.New(v.Type().Elem()).Elem()
				if err := self.decodeValue(elem); err != nil {
					return err
				}
				v.Set(reflect.Append(v, elem))
			}
			self.sigOffset = afterElemOffset
			return nil
		case v.Kind() == reflect.Map:
			if self.signature[elemSigOffset] != '{' {
				return errors.New("Expected type code '{' but got " + string(self.signature[elemSigOffset]) + " when decoding to map")
			}
			v.Set(reflect.MakeMap(v.Type()))
			for self.dataOffset < arrayEnd {
				self.align(8)
				// Reset signature offset to first
				// item in dictionary entry:
				self.sigOffset = elemSigOffset + 1
				key := reflect.New(v.Type().Key()).Elem()
				value := reflect.New(v.Type().Elem()).Elem()
				if err := self.decodeValue(key); err != nil {
					return err
				}
				if err := self.decodeValue(value); err != nil {
					return err
				}
				v.SetMapIndex(key, value)
			}
			self.sigOffset = afterElemOffset
			return nil
		case typeBlankInterface.AssignableTo(v.Type()):
			// XXX: Need to support maps here (i.e. next
			// signature char is '{')
			array := make([]interface{}, 0)
			for self.dataOffset < arrayEnd {
				// Reset signature offset to the array element.
				self.sigOffset = elemSigOffset
				var elem interface{}
				if err := self.decodeValue(reflect.ValueOf(&elem).Elem()); err != nil {
					return err
				}
				array = append(array, elem)
			}
			v.Set(reflect.ValueOf(array))
			return nil
		}
	case '(':
		self.align(8)
		// Do we have a pointer to a struct?
		if v.Kind() == reflect.Ptr && v.Type().Elem().Kind() == reflect.Struct {
			if v.IsNil() {
				v.Set(reflect.New(v.Type().Elem()))
			}
			v = v.Elem()
		}
		switch {
		case v.Kind() == reflect.Struct:
			for i := 0; i < v.NumField() && self.sigOffset < len(self.signature) && self.signature[self.sigOffset] != ')'; i++ {
				if err := self.decodeValue(v.Field(i)); err != nil {
					return err
				}
			}
			if self.sigOffset >= len(self.signature) || self.signature[self.sigOffset] != ')' {
				return signatureOverrunError
			}
			// move past the closing parentheses
			self.sigOffset += 1
			return nil
		case typeBlankInterface.AssignableTo(v.Type()):
			// Decode as a slice of interface{} values.
			s := make([]interface{}, 0)
			for self.sigOffset < len(self.signature) && self.signature[self.sigOffset] != ')' {
				var field interface{}
				if err := self.decodeValue(reflect.ValueOf(&field).Elem()); err != nil {
					return err
				}
				s = append(s, field)
			}
			v.Set(reflect.ValueOf(s))
			return nil
		}
	case 'v':
		var variant *Variant
		switch {
		case v.Kind() == reflect.Ptr && v.Type().Elem() == typeVariant:
			if v.IsNil() {
				variant = &Variant{}
				v.Set(reflect.ValueOf(variant))
			} else {
				variant = v.Interface().(*Variant)
			}
		case v.Type() == typeVariant:
			variant = v.Addr().Interface().(*Variant)
		case typeBlankInterface.AssignableTo(v.Type()):
			variant = &Variant{}
			v.Set(reflect.ValueOf(variant))
		}
		if variant != nil {
			signature, err := self.readSignature()
			if err != nil {
				return err
			}
			// Decode the variant value through a sub-decoder.
			variantDec := decoder{
				signature: signature,
				data: self.data,
				order: self.order,
				dataOffset: self.dataOffset,
				sigOffset: 0}
			if err := variantDec.decodeValue(reflect.ValueOf(&variant.Value).Elem()); err != nil {
				return err
			}
			// Decoding continues after the variant value.
			self.dataOffset = variantDec.dataOffset
			return nil
		}
	}
	return errors.New("Could not decode " + string(sigCode) + " to " + v.Type().String())
}
