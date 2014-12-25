// Copyright 2014, Hǎiliàng Wáng. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package flow

import (
	"bytes"
	"encoding"
	"fmt"
	"io"
	"math/big"
	"reflect"
	"strconv"
)

type (
	ValueEncodeFunc func(v reflect.Value, w io.Writer) error
	ValueDecodeFunc func(val []byte, v reflect.Value) error
	marshalFunc     func() (text []byte, err error)
	unmarshalFunc   func(text []byte) error
)

type ValueEncoding struct {
	Encode ValueEncodeFunc
	Decode ValueDecodeFunc
}

var typeToValueEncoding = map[reflect.Kind]ValueEncoding{
	reflect.Bool:       ValueEncoding{encodeBool, decodeBool},
	reflect.Int8:       ValueEncoding{encodeInt, decodeInt},
	reflect.Int16:      ValueEncoding{encodeInt, decodeInt},
	reflect.Int32:      ValueEncoding{encodeInt, decodeInt},
	reflect.Int64:      ValueEncoding{encodeInt, decodeInt},
	reflect.Int:        ValueEncoding{encodeInt, decodeInt},
	reflect.Uint8:      ValueEncoding{encodeUint, decodeUint},
	reflect.Uint16:     ValueEncoding{encodeUint, decodeUint},
	reflect.Uint32:     ValueEncoding{encodeUint, decodeUint},
	reflect.Uint64:     ValueEncoding{encodeUint, decodeUint},
	reflect.Uint:       ValueEncoding{encodeUint, decodeUint},
	reflect.Uintptr:    ValueEncoding{encodeUint, decodeUint},
	reflect.Float32:    ValueEncoding{encodeFloat32, decodeFloat32},
	reflect.Float64:    ValueEncoding{encodeFloat64, decodeFloat64},
	reflect.Complex64:  ValueEncoding{encodeComplex64, decodeComplex},
	reflect.Complex128: ValueEncoding{encodeComplex128, decodeComplex},
	reflect.String:     ValueEncoding{encodeString, decodeString},
}

func marshal(f marshalFunc, w io.Writer) error {
	b, err := f()
	if err != nil {
		return err
	}
	if bytes.IndexAny(b, "\t {},") != -1 {
		b = []byte(strconv.Quote(string(b)))
	}
	w.Write(b)
	return nil
}

func unmarshal(f unmarshalFunc, val []byte, v reflect.Value) error {
	s, err := strconv.Unquote(string(val))
	if err != nil {
		return f(val)
	}
	return f([]byte(s))
}

func encodeMarshaler(v reflect.Value, w io.Writer) error {
	return marshal(v.Interface().(Marshaler).MarshalOGDL, w)
}

func decodeMarshaler(val []byte, v reflect.Value) error {
	return unmarshal(v.Interface().(Unmarshaler).UnmarshalOGDL, val, v)
}

func encodeTextMarshaler(v reflect.Value, w io.Writer) error {
	return marshal(v.Interface().(encoding.TextMarshaler).MarshalText, w)
}

func decodeTextMarshaler(val []byte, v reflect.Value) error {
	return unmarshal(v.Interface().(encoding.TextUnmarshaler).UnmarshalText, val, v)
}

func encodeBool(v reflect.Value, w io.Writer) error {
	return writeString(w, strconv.FormatBool(v.Bool()))
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

func encodeInt(v reflect.Value, w io.Writer) error {
	return writeString(w, strconv.FormatInt(v.Int(), 10))
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

func encodeUint(v reflect.Value, w io.Writer) error {
	return writeString(w, strconv.FormatUint(v.Uint(), 10))
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

func encodeFloat32(v reflect.Value, w io.Writer) error {
	return encodeFloat(v, w, 32)
}

func decodeFloat32(val []byte, v reflect.Value) error {
	return decodeFloat(val, v, 32)
}

func encodeFloat64(v reflect.Value, w io.Writer) error {
	return encodeFloat(v, w, 32)
}

func decodeFloat64(val []byte, v reflect.Value) error {
	return decodeFloat(val, v, 64)
}

func encodeFloat(v reflect.Value, w io.Writer, bit int) error {
	return writeString(w, strconv.FormatFloat(v.Float(), 'g', -1, bit))
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

func encodeComplex64(v reflect.Value, w io.Writer) error {
	return encodeComplex(v, w, 32)
}

func encodeComplex128(v reflect.Value, w io.Writer) error {
	return encodeComplex(v, w, 64)
}

func encodeComplex(v reflect.Value, w io.Writer, bitSize int) error {
	c := v.Complex()
	r, i := real(c), imag(c)
	if err := writeString(w, strconv.FormatFloat(r, 'g', -1, bitSize)); err != nil {
		return err
	}
	if i >= 0 {
		if err := writeByte(w, '+'); err != nil {
			return err
		}
	}
	if err := writeString(w, strconv.FormatFloat(i, 'g', -1, bitSize)); err != nil {
		return err
	}
	if err := writeByte(w, 'i'); err != nil {
		return err
	}
	return nil
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

func encodeString(v reflect.Value, w io.Writer) error {
	return writeString(w, strconv.Quote(v.String()))
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
