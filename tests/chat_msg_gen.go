package tests

// Code generated by github.com/tinylib/msgp DO NOT EDIT.

import (
	"github.com/tinylib/msgp/msgp"
)

// DecodeMsg implements msgp.Decodable
func (z *ChatMsg_Login) DecodeMsg(dc *msgp.Reader) (err error) {
	var field []byte
	_ = field
	var zb0001 uint32
	zb0001, err = dc.ReadMapHeader()
	if err != nil {
		err = msgp.WrapError(err)
		return
	}
	for zb0001 > 0 {
		zb0001--
		field, err = dc.ReadMapKeyPtr()
		if err != nil {
			err = msgp.WrapError(err)
			return
		}
		switch msgp.UnsafeString(field) {
		case "name":
			z.Name, err = dc.ReadString()
			if err != nil {
				err = msgp.WrapError(err, "Name")
				return
			}
		default:
			err = dc.Skip()
			if err != nil {
				err = msgp.WrapError(err)
				return
			}
		}
	}
	return
}

// EncodeMsg implements msgp.Encodable
func (z ChatMsg_Login) EncodeMsg(en *msgp.Writer) (err error) {
	// map header, size 1
	// write "name"
	err = en.Append(0x81, 0xa4, 0x6e, 0x61, 0x6d, 0x65)
	if err != nil {
		return
	}
	err = en.WriteString(z.Name)
	if err != nil {
		err = msgp.WrapError(err, "Name")
		return
	}
	return
}

// MarshalMsg implements msgp.Marshaler
func (z ChatMsg_Login) MarshalMsg(b []byte) (o []byte, err error) {
	o = msgp.Require(b, z.Msgsize())
	// map header, size 1
	// string "name"
	o = append(o, 0x81, 0xa4, 0x6e, 0x61, 0x6d, 0x65)
	o = msgp.AppendString(o, z.Name)
	return
}

// UnmarshalMsg implements msgp.Unmarshaler
func (z *ChatMsg_Login) UnmarshalMsg(bts []byte) (o []byte, err error) {
	var field []byte
	_ = field
	var zb0001 uint32
	zb0001, bts, err = msgp.ReadMapHeaderBytes(bts)
	if err != nil {
		err = msgp.WrapError(err)
		return
	}
	for zb0001 > 0 {
		zb0001--
		field, bts, err = msgp.ReadMapKeyZC(bts)
		if err != nil {
			err = msgp.WrapError(err)
			return
		}
		switch msgp.UnsafeString(field) {
		case "name":
			z.Name, bts, err = msgp.ReadStringBytes(bts)
			if err != nil {
				err = msgp.WrapError(err, "Name")
				return
			}
		default:
			bts, err = msgp.Skip(bts)
			if err != nil {
				err = msgp.WrapError(err)
				return
			}
		}
	}
	o = bts
	return
}

// Msgsize returns an upper bound estimate of the number of bytes occupied by the serialized message
func (z ChatMsg_Login) Msgsize() (s int) {
	s = 1 + 5 + msgp.StringPrefixSize + len(z.Name)
	return
}

// DecodeMsg implements msgp.Decodable
func (z *ChatMsg_Send) DecodeMsg(dc *msgp.Reader) (err error) {
	var field []byte
	_ = field
	var zb0001 uint32
	zb0001, err = dc.ReadMapHeader()
	if err != nil {
		err = msgp.WrapError(err)
		return
	}
	for zb0001 > 0 {
		zb0001--
		field, err = dc.ReadMapKeyPtr()
		if err != nil {
			err = msgp.WrapError(err)
			return
		}
		switch msgp.UnsafeString(field) {
		case "recv":
			z.Reciever, err = dc.ReadString()
			if err != nil {
				err = msgp.WrapError(err, "Reciever")
				return
			}
		case "msg":
			z.Content, err = dc.ReadString()
			if err != nil {
				err = msgp.WrapError(err, "Content")
				return
			}
		default:
			err = dc.Skip()
			if err != nil {
				err = msgp.WrapError(err)
				return
			}
		}
	}
	return
}

// EncodeMsg implements msgp.Encodable
func (z ChatMsg_Send) EncodeMsg(en *msgp.Writer) (err error) {
	// map header, size 2
	// write "recv"
	err = en.Append(0x82, 0xa4, 0x72, 0x65, 0x63, 0x76)
	if err != nil {
		return
	}
	err = en.WriteString(z.Reciever)
	if err != nil {
		err = msgp.WrapError(err, "Reciever")
		return
	}
	// write "msg"
	err = en.Append(0xa3, 0x6d, 0x73, 0x67)
	if err != nil {
		return
	}
	err = en.WriteString(z.Content)
	if err != nil {
		err = msgp.WrapError(err, "Content")
		return
	}
	return
}

// MarshalMsg implements msgp.Marshaler
func (z ChatMsg_Send) MarshalMsg(b []byte) (o []byte, err error) {
	o = msgp.Require(b, z.Msgsize())
	// map header, size 2
	// string "recv"
	o = append(o, 0x82, 0xa4, 0x72, 0x65, 0x63, 0x76)
	o = msgp.AppendString(o, z.Reciever)
	// string "msg"
	o = append(o, 0xa3, 0x6d, 0x73, 0x67)
	o = msgp.AppendString(o, z.Content)
	return
}

// UnmarshalMsg implements msgp.Unmarshaler
func (z *ChatMsg_Send) UnmarshalMsg(bts []byte) (o []byte, err error) {
	var field []byte
	_ = field
	var zb0001 uint32
	zb0001, bts, err = msgp.ReadMapHeaderBytes(bts)
	if err != nil {
		err = msgp.WrapError(err)
		return
	}
	for zb0001 > 0 {
		zb0001--
		field, bts, err = msgp.ReadMapKeyZC(bts)
		if err != nil {
			err = msgp.WrapError(err)
			return
		}
		switch msgp.UnsafeString(field) {
		case "recv":
			z.Reciever, bts, err = msgp.ReadStringBytes(bts)
			if err != nil {
				err = msgp.WrapError(err, "Reciever")
				return
			}
		case "msg":
			z.Content, bts, err = msgp.ReadStringBytes(bts)
			if err != nil {
				err = msgp.WrapError(err, "Content")
				return
			}
		default:
			bts, err = msgp.Skip(bts)
			if err != nil {
				err = msgp.WrapError(err)
				return
			}
		}
	}
	o = bts
	return
}

// Msgsize returns an upper bound estimate of the number of bytes occupied by the serialized message
func (z ChatMsg_Send) Msgsize() (s int) {
	s = 1 + 5 + msgp.StringPrefixSize + len(z.Reciever) + 4 + msgp.StringPrefixSize + len(z.Content)
	return
}