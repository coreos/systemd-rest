package dbus

import (
	"encoding/binary"
	"errors"
	"io"
)

// See the D-Bus tutorial for information about message types.
//		http://dbus.freedesktop.org/doc/dbus-tutorial.html#messages
type MessageType uint8

const (
	TypeInvalid MessageType = iota
	TypeMethodCall
	TypeMethodReturn
	TypeError
	TypeSignal
)

var messageTypeString = map[MessageType]string{
	TypeInvalid:      "invalid",
	TypeMethodCall:   "method_call",
	TypeMethodReturn: "method_return",
	TypeSignal:       "signal",
	TypeError:        "error",
}

func (t MessageType) String() string { return messageTypeString[t] }

type MessageFlag uint8

const (
	FlagNoReplyExpected MessageFlag = 1 << iota
	FlagNoAutoStart
)

type Message struct {
	order       binary.ByteOrder
	Type        MessageType
	Flags       MessageFlag
	Protocol    uint8
	bodyLength  int
	serial      uint32
	// header fields
	Path        ObjectPath
	Iface       string
	Member      string
	ErrorName   string
	replySerial uint32
	Dest        string
	Sender      string
	sig         Signature
	body        []byte
}

// Create a new message with Flags == 0 and Protocol == 1.
func newMessage() *Message {
	msg := new(Message)

	msg.order = binary.LittleEndian
	msg.serial = 0
	msg.replySerial = 0
	msg.Flags = 0
	msg.Protocol = 1

	msg.body = []byte{}

	return msg
}

func NewMethodCallMessage(destination string, path ObjectPath, iface string, member string) *Message {
	msg := newMessage()
	msg.Type = TypeMethodCall
	msg.Dest = destination
	msg.Path = path
	msg.Iface = iface
	msg.Member = member
	return msg
}

func NewMethodReturnMessage(methodCall *Message) *Message {
	if methodCall.serial == 0 {
		panic("methodCall.serial == 0")
	}
	if methodCall.Type != TypeMethodCall {
		panic("replies should be sent in response to method calls")
	}
	msg := newMessage()
	msg.Type = TypeMethodReturn
	msg.replySerial = methodCall.serial
	msg.Dest = methodCall.Sender
	return msg
}

func NewSignalMessage(path ObjectPath, iface string, member string) *Message {
	msg := newMessage()
	msg.Type = TypeSignal
	msg.Path = path
	msg.Iface = iface
	msg.Member = member
	return msg
}

func NewErrorMessage(methodCall *Message, errorName string, message string) *Message {
	if methodCall.serial == 0 {
		panic("methodCall.serial == 0")
	}
	if methodCall.Type != TypeMethodCall {
		panic("errors should be sent in response to method calls")
	}
	msg := newMessage()
	msg.Type = TypeError
	msg.replySerial = methodCall.serial
	msg.Dest = methodCall.Sender
	msg.ErrorName = errorName
	if err := msg.AppendArgs(message); err != nil {
		panic(err)
	}
	return msg
}

func (p *Message) setSerial(serial uint32) {
	if p.serial != 0 {
		panic("Message already has a serial number")
	}
	p.serial = serial
}

func (p *Message) AppendArgs(args ...interface{}) error {
	enc := newEncoder(p.sig, p.body, p.order)
	if err := enc.Append(args...); err != nil {
		return err
	}
	p.sig = enc.signature
	p.body = enc.data.Bytes()
	return nil
}

func (p *Message) GetArgs(args ...interface{}) error {
	dec := newDecoder(p.sig, p.body, p.order)
	return dec.Decode(args...)
}

func (p *Message) GetAllArgs() []interface{} {
	dec := newDecoder(p.sig, p.body, p.order)
	args := make([]interface{}, 0)
	for dec.HasMore() {
		var arg interface{}
		if err := dec.Decode(&arg); err != nil {
			panic(err)
		}
		args = append(args, arg)
	}
	return args
}

func (p *Message) AsError() error {
	if p.Type != TypeError {
		panic("Only messages of type 'error' can be converted to an error")
	}
	var errorMessage string
	if err := p.GetArgs(&errorMessage); err != nil {
		// Ignore error
		errorMessage = ""
	}
	return &Error{p.ErrorName, errorMessage}
}

type headerField struct {
	Code byte
	Value Variant
}

func readMessage(r io.Reader) (*Message, error) {
	header := make([]byte, 16)
	if n, err := r.Read(header); n < len(header) {
		if err == nil {
			err = errors.New("Could not read message header")
		}
		return nil, err
	}

	msg := newMessage()
	switch header[0] {
	case 'l':
		msg.order = binary.LittleEndian
	case 'B':
		msg.order = binary.BigEndian
	default:
		return nil, errors.New("Unknown message endianness: " + string(header[0]))
	}
	dec := newDecoder("yyyyuuu", header, msg.order)
	var msgOrder byte
	var msgBodyLength, headerFieldsLength uint32
	if err := dec.Decode(&msgOrder, &msg.Type, &msg.Flags, &msg.Protocol, &msgBodyLength, &msg.serial, &headerFieldsLength); err != nil {
		return nil,  err
	}

	// Read out and decode the header fields, plus the padding to
	// 8 bytes.
	padding := -(len(header) + int(headerFieldsLength)) % 8
	if padding < 0 {
		padding += 8
	}
	headerFields := make([]byte, 16 + int(headerFieldsLength) + padding)
	copy(headerFields[:16], header)
	if n, err := r.Read(headerFields[16:]); n < len(headerFields) - 16 {
		if err == nil {
			err = errors.New("Could not read message header fields")
		}
		return nil, err
	}
	dec = newDecoder("a(yv)", headerFields, msg.order)
	dec.dataOffset += 12
	fields := make([]headerField, 0, 10)
	if err := dec.Decode(&fields); err != nil {
		return nil,  err
	}
	for _, field := range fields {
		switch field.Code {
		case 1:
			msg.Path = field.Value.Value.(ObjectPath)
		case 2:
			msg.Iface = field.Value.Value.(string)
		case 3:
			msg.Member = field.Value.Value.(string)
		case 4:
			msg.ErrorName = field.Value.Value.(string)
		case 5:
			msg.replySerial = field.Value.Value.(uint32)
		case 6:
			msg.Dest = field.Value.Value.(string)
		case 7:
			msg.Sender = field.Value.Value.(string)
		case 8:
			msg.sig = field.Value.Value.(Signature)
		}
	}

	msg.body = make([]byte, msgBodyLength)
	if n, err := r.Read(msg.body); n < len(msg.body) {
		if err == nil {
			err = errors.New("Could not read message body")
		}
		return nil, err
	}
	return msg, nil
}

func (p *Message) WriteTo(w io.Writer) (int64, error) {
	fields := make([]headerField, 0, 10)
	if p.Path != "" {
		fields = append(fields, headerField{1, Variant{p.Path}})
	}
	if p.Iface != "" {
		fields = append(fields, headerField{2, Variant{p.Iface}})
	}
	if p.Member != "" {
		fields = append(fields, headerField{3, Variant{p.Member}})
	}
	if p.ErrorName != "" {
		fields = append(fields, headerField{4, Variant{p.ErrorName}})
	}
	if p.replySerial != 0 {
		fields = append(fields, headerField{5, Variant{p.replySerial}})
	}
	if p.Dest != "" {
		fields = append(fields, headerField{6, Variant{p.Dest}})
	}
	if p.Sender != "" {
		fields = append(fields, headerField{7, Variant{p.Sender}})
	}
	if p.sig != "" {
		fields = append(fields, headerField{8, Variant{p.sig}})
	}

	var orderTag byte
	switch p.order {
	case binary.LittleEndian:
		orderTag = 'l'
	case binary.BigEndian:
		orderTag = 'B'
	default:
		return 0, errors.New("Unknown byte order: " + p.order.String())
	}

	header := newEncoder("", nil, p.order)
	if err := header.Append(orderTag, byte(p.Type), byte(p.Flags), byte(p.Protocol), uint32(len(p.body)), p.serial, fields); err != nil {
		return 0, err
	}

	// Add alignment bytes for body
	header.align(8)
	m, err := w.Write(header.data.Bytes())
	if err != nil {
		return int64(m), err
	} else if m != header.data.Len() {
		return int64(m), errors.New("Failed to write complete message header")
	}

	n, err := w.Write(p.body)
	if err != nil {
		return int64(m + n), err
	} else if n != len(p.body) {
		return int64(m + n), errors.New("Failed to write complete message body")
	}
	return int64(m + n), nil
}
