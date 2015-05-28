// Copyright 2014, Hǎiliàng Wáng. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package tff

import (
	"encoding"
	"fmt"
	"io"
	"math/big"
	"reflect"
	"strconv"
	"strings"
)

type (
	BasicEncodeFunc func(v reflect.Value) Node
	BasicDecodeFunc func(val []byte, v reflect.Value) error
	marshalFunc     func() (_ []byte, err error)
	unmarshalFunc   func(text []byte) error
)

type BasicEncoding struct {
	Encode BasicEncodeFunc
	Decode BasicDecodeFunc
}

var kindToBasicEncoding = map[reflect.Kind]BasicEncoding{
	reflect.Bool:       BasicEncoding{encodeBool, decodeBool},
	reflect.Int8:       BasicEncoding{encodeInt, decodeInt},
	reflect.Int16:      BasicEncoding{encodeInt, decodeInt},
	reflect.Int32:      BasicEncoding{encodeInt, decodeInt},
	reflect.Int64:      BasicEncoding{encodeInt, decodeInt},
	reflect.Int:        BasicEncoding{encodeInt, decodeInt},
	reflect.Uint8:      BasicEncoding{encodeUint, decodeUint},
	reflect.Uint16:     BasicEncoding{encodeUint, decodeUint},
	reflect.Uint32:     BasicEncoding{encodeUint, decodeUint},
	reflect.Uint64:     BasicEncoding{encodeUint, decodeUint},
	reflect.Uint:       BasicEncoding{encodeUint, decodeUint},
	reflect.Uintptr:    BasicEncoding{encodeUint, decodeUint},
	reflect.Float32:    BasicEncoding{encodeFloat32, decodeFloat32},
	reflect.Float64:    BasicEncoding{encodeFloat64, decodeFloat64},
	reflect.Complex64:  BasicEncoding{encodeComplex64, decodeComplex},
	reflect.Complex128: BasicEncoding{encodeComplex128, decodeComplex},
	reflect.String:     BasicEncoding{encodeString, decodeString},
}

func marshal(f marshalFunc) Node {
	buf, err := f()
	if err != nil {
		panic(err)
	}
	node := Node{string(buf), nil}
	if strings.IndexAny(node.Value, "\r\n") != -1 {
		node.Value = strconv.Quote(node.Value)
	}
	return node
}

func unmarshal(f unmarshalFunc, val []byte, v reflect.Value) error {
	s, err := strconv.Unquote(string(val))
	if err != nil {
		return f(val)
	}
	return f([]byte(s))
}

func encodeMarshaler(v reflect.Value) Node {
	return marshal(v.Interface().(Marshaler).MarshalTFF)
}

func decodeMarshaler(val []byte, v reflect.Value) error {
	return unmarshal(v.Interface().(Unmarshaler).UnmarshalTFF, val, v)
}

func encodeTextMarshaler(v reflect.Value) Node {
	return marshal(v.Interface().(encoding.TextMarshaler).MarshalText)
}

func decodeTextMarshaler(val []byte, v reflect.Value) error {
	return unmarshal(v.Interface().(encoding.TextUnmarshaler).UnmarshalText, val, v)
}

func encodeBool(v reflect.Value) Node {
	return Node{strconv.FormatBool(v.Bool()), nil}
}

func decodeBool(val []byte, v reflect.Value) error {
	switch string(val) {
	case "true":
		v.SetBool(true)
	case "false":
		v.SetBool(false)
	default:
		return fmt.Errorf("unexpected bool value: %s", strconv.Quote(string(val)))
	}
	return nil
}

func encodeInt(v reflect.Value) Node {
	return Node{strconv.FormatInt(v.Int(), 10), nil}
}

func decodeInt(val []byte, v reflect.Value) error {
	i := big.NewInt(0)
	i, ok := i.SetString(string(val), 10)
	if !ok {
		return fmt.Errorf("unexpected int value: %s", strconv.Quote(string(val)))
	}
	// TODO: handle overflow
	v.SetInt(i.Int64())
	return nil
}

func encodeUint(v reflect.Value) Node {
	return Node{strconv.FormatUint(v.Uint(), 10), nil}
}

func decodeUint(val []byte, v reflect.Value) error {
	i := big.NewInt(0)
	i, ok := i.SetString(string(val), 10)
	if !ok {
		return fmt.Errorf("unexpected int value: %s", strconv.Quote(string(val)))
	}
	// TODO: handle overflow
	v.SetUint(i.Uint64())
	return nil
}

func encodeFloat32(v reflect.Value) Node {
	return encodeFloat(v, 32)
}

func decodeFloat32(val []byte, v reflect.Value) error {
	return decodeFloat(val, v, 32)
}

func encodeFloat64(v reflect.Value) Node {
	return encodeFloat(v, 32)
}

func decodeFloat64(val []byte, v reflect.Value) error {
	return decodeFloat(val, v, 64)
}

func encodeFloat(v reflect.Value, bit int) Node {
	return Node{strconv.FormatFloat(v.Float(), 'g', -1, bit), nil}
}

func decodeFloat(val []byte, v reflect.Value, bit int) error {
	f, err := strconv.ParseFloat(string(val), bit)
	if err != nil {
		return fmt.Errorf("unexpected float value: %s", strconv.Quote(string(val)))
	}
	// TODO: handle overflow
	v.SetFloat(f)
	return nil
}

func encodeComplex64(v reflect.Value) Node {
	return encodeComplex(v, 32)
}

func encodeComplex128(v reflect.Value) Node {
	return encodeComplex(v, 64)
}

func encodeComplex(v reflect.Value, bitSize int) Node {
	c := v.Complex()
	r, i := real(c), imag(c)
	if i >= 0 {
		return Node{fmt.Sprintf("%s+%si",
			strconv.FormatFloat(r, 'g', -1, bitSize),
			strconv.FormatFloat(i, 'g', -1, bitSize)),
			nil}
	}
	return Node{fmt.Sprintf("%s%si",
		strconv.FormatFloat(r, 'g', -1, bitSize),
		strconv.FormatFloat(i, 'g', -1, bitSize)),
		nil}
}

func decodeComplex(val []byte, v reflect.Value) error {
	var c complex128
	if _, err := fmt.Sscan(string(val), &c); err != nil {
		return err
	}
	// TODO: handle overflow
	v.SetComplex(c)
	return nil
}

func encodeString(v reflect.Value) Node {
	return Node{strconv.Quote(v.String()), nil}
}

func decodeString(val []byte, v reflect.Value) error {
	s, err := strconv.Unquote(string(val))
	if err != nil {
		s = string(val)
	}
	v.SetString(s)
	return nil
}

func writeString(w io.Writer, s string) error {
	_, err := w.Write([]byte(s))
	return err
}

func writeByte(w io.Writer, b byte) error {
	_, err := w.Write([]byte{b})
	return err
}
