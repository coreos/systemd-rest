package dbus

import (
	"errors"
	"fmt"
	"reflect"
)

var (
	typeHasObjectPath = reflect.TypeOf((*HasObjectPath)(nil)).Elem()
	typeVariant = reflect.TypeOf(Variant{})
	typeSignature = reflect.TypeOf(Signature(""))
	typeBlankInterface = reflect.TypeOf((*interface{})(nil)).Elem()
)


type Signature string

func SignatureOf(t reflect.Type) (Signature, error) {
	if t.AssignableTo(typeHasObjectPath) {
		return Signature("o"), nil
	}
	switch t.Kind() {
	case reflect.Uint8:
		return Signature("y"), nil
	case reflect.Bool:
		return Signature("b"), nil
	case reflect.Int16:
		return Signature("n"), nil
	case reflect.Uint16:
		return Signature("q"), nil
	case reflect.Int32:
		return Signature("i"), nil
	case reflect.Uint32:
		return Signature("u"), nil
	case reflect.Int64:
		return Signature("x"), nil
	case reflect.Uint64:
		return Signature("t"), nil
	case reflect.Float64:
		return Signature("d"), nil
	case reflect.String:
		if t == typeSignature {
			return Signature("g"), nil
		}
		return Signature("s"), nil
	case reflect.Array, reflect.Slice:
		valueSig, err := SignatureOf(t.Elem())
		if err != nil {
			return Signature(""), err
		}
		return Signature("a") + valueSig, nil
	case reflect.Map:
		keySig, err := SignatureOf(t.Key())
		if err != nil {
			return Signature(""), err
		}
		valueSig, err := SignatureOf(t.Elem())
		if err != nil {
			return Signature(""), err
		}
		return Signature("a{") + keySig + valueSig + Signature("}"), nil
	case reflect.Struct:
		// Special case the variant structure
		if t == typeVariant {
			return Signature("v"), nil
		}

		sig := Signature("(")
		for i := 0; i != t.NumField(); i++ {
			fieldSig, err := SignatureOf(t.Field(i).Type)
			if err != nil {
				return Signature(""), err
			}
			sig += fieldSig
		}
		sig += Signature(")")
		return sig, nil
	case reflect.Ptr:
		// dereference pointers
		sig, err := SignatureOf(t.Elem())
		return sig, err
	}
	return Signature(""), errors.New("Can not determine signature for " + t.String())
}

func (sig Signature) NextType(offset int) (next int, err error) {
	if offset >= len(sig) {
		err = errors.New("No more types codes in signature")
		return
	}
	switch sig[offset] {
	case 'y', 'b', 'n', 'q', 'i', 'u', 'x', 't', 'd', 's', 'o', 'g', 'v', 'h':
		// A basic type code.
		next = offset + 1
	case 'a':
		// An array: consume the embedded type code
		next, err = sig.NextType(offset + 1)
	case '{':
		// A pair used in maps: consume the two contained types
		next, err = sig.NextType(offset + 1)
		if err != nil {
			return
		}
		next, err = sig.NextType(next)
		if err != nil {
			return
		}
		if next >= len(sig) || sig[next] != '}' {
			err = errors.New("Pair does not end with '}'")
			return
		}
		next += 1
	case '(':
		// A struct: consume types until we 
		next = offset + 1
		for {
			if next < len(sig) && sig[next] == ')' {
				next += 1
				return
			}
			next, err = sig.NextType(next)
			if err != nil {
				return
			}
		}
	default:
		err = errors.New("Unknown type code " + string(sig[offset]))
	}
	return
}

// Validate that the signature is a valid string of type codes
func (sig Signature) Validate() (err error) {
	offset := 0
	for offset < len(sig) {
		offset, err = sig.NextType(offset)
		if err != nil {
			break
		}
	}
	return
}


type ObjectPath string

type HasObjectPath interface {
	GetObjectPath() ObjectPath
}

func (o ObjectPath) GetObjectPath() ObjectPath {
	return o
}

type Variant struct {
	Value interface{}
}

func (v *Variant) GetVariantSignature() (Signature, error) {
	return SignatureOf(reflect.TypeOf(v.Value))
}


type Error struct {
	Name string
	Message string
}

func (e *Error) Error() string {
	return fmt.Sprint(e.Name, ": ", e.Message)
}
