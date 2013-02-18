package dbus

import (
	"bytes"
	"encoding/binary"
	"errors"
	"reflect"
)

type encoder struct {
	signature Signature
	data bytes.Buffer
	order binary.ByteOrder
}

func newEncoder(signature Signature, data []byte, order binary.ByteOrder) *encoder {
	enc := &encoder{signature: signature, order: order}
	if data != nil {
		enc.data.Write(data)
	}
	return enc
}

func (self *encoder) align(alignment int) {
	for self.data.Len() % alignment != 0 {
		self.data.WriteByte(0)
	}
}

func (self *encoder) Append(args ...interface{}) error {
	for _, arg := range args {
		if err := self.appendValue(reflect.ValueOf(arg)); err != nil {
			return err
		}
	}
	return nil
}

func (self *encoder) alignForType(t reflect.Type) error {
	// If type matches the HasObjectPath interface, treat like an
	// ObjectPath
	if t.AssignableTo(typeHasObjectPath) {
		t = reflect.TypeOf(ObjectPath(""))
	}
	// dereference pointers
	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	switch t.Kind() {
	case reflect.Uint8:
		self.align(1)
	case reflect.Int16, reflect.Uint16:
		self.align(2)
	case reflect.Bool, reflect.Int32, reflect.Uint32, reflect.Array, reflect.Slice, reflect.Map:
		self.align(4)
	case reflect.Int64, reflect.Uint64, reflect.Float64:
		self.align(8)
	case reflect.String:
		if t == typeSignature {
			self.align(1)
		} else {
			self.align(4)
		}
	case reflect.Struct:
		if t == typeVariant {
			self.align(1)
		} else {
			self.align(8)
		}
	default:
		return errors.New("Don't know how to align " + t.String())
	}
	return nil
}

func (self *encoder) appendValue(v reflect.Value) error {
	signature, err := SignatureOf(v.Type())
	if err != nil {
		return err
	}
	self.signature += signature

	// Convert HasObjectPath values to ObjectPath strings
	if v.Type().AssignableTo(typeHasObjectPath) {
		path := v.Interface().(HasObjectPath).GetObjectPath()
		v = reflect.ValueOf(path)
	}

	// We want pointer values here, rather than the pointers themselves.
	for v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	self.alignForType(v.Type())
	switch v.Kind() {
	case reflect.Uint8:
		self.data.WriteByte(byte(v.Uint()))
		return nil
	case reflect.Bool:
		var uintval uint32
		if v.Bool() {
			uintval = 1
		} else {
			uintval = 0
		}
		binary.Write(&self.data, self.order, uintval)
		return nil
	case reflect.Int16:
		binary.Write(&self.data, self.order, int16(v.Int()))
		return nil
	case reflect.Uint16:
		binary.Write(&self.data, self.order, uint16(v.Uint()))
		return nil
	case reflect.Int32:
		binary.Write(&self.data, self.order, int32(v.Int()))
		return nil
	case reflect.Uint32:
		binary.Write(&self.data, self.order, uint32(v.Uint()))
		return nil
	case reflect.Int64:
		binary.Write(&self.data, self.order, int64(v.Int()))
		return nil
	case reflect.Uint64:
		binary.Write(&self.data, self.order, uint64(v.Uint()))
		return nil
	case reflect.Float64:
		binary.Write(&self.data, self.order, float64(v.Float()))
		return nil
	case reflect.String:
		s := v.String()
		// Signatures only use a single byte for the length.
		if v.Type() == typeSignature {
			self.data.WriteByte(byte(len(s)))
		} else {
			binary.Write(&self.data, self.order, uint32(len(s)))
		}
		self.data.Write([]byte(s))
		self.data.WriteByte(0)
		return nil
	case reflect.Array, reflect.Slice:
		// Marshal array contents to a separate buffer so we
		// can find its length.
		var content encoder
		content.order = self.order
		for i := 0; i < v.Len(); i++ {
			if err := content.appendValue(v.Index(i)); err != nil {
				return err
			}
		}
		binary.Write(&self.data, self.order, uint32(content.data.Len()))
		self.alignForType(v.Type().Elem())
		self.data.Write(content.data.Bytes())
		return nil
	case reflect.Map:
		// Marshal array contents to a separate buffer so we
		// can find its length.
		var content encoder
		content.order = self.order
		for _, key := range v.MapKeys() {
			content.align(8)
			if err := content.appendValue(key); err != nil {
				return err
			}
			if err := content.appendValue(v.MapIndex(key)); err != nil {
				return err
			}
		}
		binary.Write(&self.data, self.order, uint32(content.data.Len()))
		self.align(8) // alignment of DICT_ENTRY
		self.data.Write(content.data.Bytes())
		return nil
	case reflect.Struct:
		if v.Type() == typeVariant {
			variant := v.Interface().(Variant)
			variantSig, err := variant.GetVariantSignature()
			if err != nil {
				return err
			}
			// Save the signature, so we don't add the
			// typecodes for the variant value to the
			// signature.
			savedSig := self.signature
			if err := self.appendValue(reflect.ValueOf(variantSig)); err != nil {
				return err
			}
			if err := self.appendValue(reflect.ValueOf(variant.Value)); err != nil {
				return err
			}
			self.signature = savedSig
			return nil
		}
		// XXX: save and restore the signature, since we wrote
		// out the entire struct signature previously.
		savedSig := self.signature
		for i := 0; i != v.NumField(); i++ {
			if err := self.appendValue(v.Field(i)); err != nil {
				return err
			}
		}
		self.signature = savedSig
		return nil
	}
	return errors.New("Could not marshal " + v.Type().String())
}


