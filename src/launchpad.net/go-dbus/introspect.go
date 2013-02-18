package dbus

import (
	"bytes"
	"encoding/xml"
	"strings"
)

type annotationData struct {
	Name  string `xml:"name,attr"`
	Value string `xml:"value,attr"`
}

type argData struct {
	Name      string `xml:"name,attr"`
	Type      string `xml:"type,attr"`
	Direction string `xml:"direction,attr"`
}

type methodData struct {
	Name       string         `xml:"name,attr"`
	Arg        []argData      `xml:"arg"`
	Annotation annotationData `xml:"annotation"`
}

type signalData struct {
	Name string    `xml:"name,attr"`
	Arg  []argData `xml:"arg"`
}

type interfaceData struct {
	Name   string       `xml:"name,attr"`
	Method []methodData `xml:"method"`
	Signal []signalData `xml:"signal"`
}

type introspect struct {
	Name      string          `xml:"name,attr"`
	Interface []interfaceData `xml:"interface"`
	Node      []*Introspect   `xml:"node"`
}

type Introspect interface {
	GetInterfaceData(name string) InterfaceData
}

type InterfaceData interface {
	GetMethodData(name string) MethodData
	GetSignalData(name string) SignalData
	GetName() string
}

type MethodData interface {
	GetName() string
	GetInSignature() Signature
	GetOutSignature() Signature
}

type SignalData interface {
	GetName() string
	GetSignature() Signature
}

func NewIntrospect(xmlIntro string) (Introspect, error) {
	intro := new(introspect)
	buff := bytes.NewBufferString(xmlIntro)
	err := xml.Unmarshal(buff.Bytes(), intro)
	if err != nil {
		return nil, err
	}

	return intro, nil
}

func (p introspect) GetInterfaceData(name string) InterfaceData {
	for _, v := range p.Interface {
		if v.Name == name {
			return v
		}
	}
	return nil
}

func (p interfaceData) GetMethodData(name string) MethodData {
	for _, v := range p.Method {
		if v.GetName() == name {
			return v
		}
	}
	return nil
}

func (p interfaceData) GetSignalData(name string) SignalData {
	for _, v := range p.Signal {
		if v.GetName() == name {
			return v
		}
	}
	return nil
}

func (p interfaceData) GetName() string { return p.Name }

func (p methodData) GetInSignature() (sig Signature) {
	for _, v := range p.Arg {
		if strings.ToUpper(v.Direction) == "IN" {
			sig += Signature(v.Type)
		}
	}
	return
}

func (p methodData) GetOutSignature() (sig Signature) {
	for _, v := range p.Arg {
		if strings.ToUpper(v.Direction) == "OUT" {
			sig += Signature(v.Type)
		}
	}
	return
}

func (p methodData) GetName() string { return p.Name }

func (p signalData) GetSignature() (sig Signature) {
	for _, v := range p.Arg {
		sig += Signature(v.Type)
	}
	return
}

func (p signalData) GetName() string { return p.Name }
