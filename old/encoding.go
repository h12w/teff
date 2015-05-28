// Copyright 2014, Hǎiliàng Wáng. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package tff

import (
	"encoding"
	"reflect"
	"sort"
)

type Marshaler interface {
	MarshalTFF() ([]byte, error)
}

type Unmarshaler interface {
	UnmarshalTFF([]byte) error
}

type MatchFunc func(v reflect.Value) (*Encoding, bool)

type (
	EncodeFunc func(c Composer) error
	DecodeFunc func(parser Parser) error
)

type Encoding struct {
	Encode EncodeFunc
	Decode DecodeFunc
}

var (
	matchFuncs          []MatchFunc
	marshalerType       = reflect.TypeOf(new(Marshaler)).Elem()
	unmarshalerType     = reflect.TypeOf(new(Unmarshaler)).Elem()
	textMarshalerType   = reflect.TypeOf(new(encoding.TextMarshaler)).Elem()
	textUnmarshalerType = reflect.TypeOf(new(encoding.TextUnmarshaler)).Elem()
)

func init() {
	matchFuncs = []MatchFunc{
		matchMarshaler,
		matchTextMarshaler,
		matchBasic,
		matchStruct,
		matchSlice,
		matchMap,
		matchArray,
	}
}

func matchBasic(v reflect.Value) (*Encoding, bool) {
	if basicEncoding, ok := kindToBasicEncoding[v.Kind()]; ok {
		return &Encoding{
			basicEncoding.ToEncode(v),
			basicEncoding.ToDecode(v),
		}, true
	}
	return nil, false
}

func (e BasicEncoding) ToEncode(v reflect.Value) EncodeFunc {
	return func(c Composer) error {
		node := e.Encode(v)
		c.Write([]byte(node.Value))
		return nil
	}
}

func (e BasicEncoding) ToDecode(v reflect.Value) DecodeFunc {
	return func(parser Parser) error {
		val, err := parser.Value()
		if err != nil {
			return err
		}
		return e.Decode(val, v)
	}
}

func (e BasicEncoding) ToEncoding(v reflect.Value) *Encoding {
	return &Encoding{e.ToEncode(v), e.ToDecode(v)}
}

func matchMarshaler(v reflect.Value) (*Encoding, bool) {
	var m Marshaler
	var u Unmarshaler
	if v.Type().Implements(marshalerType) {
		m = v.Interface().(Marshaler)
	}
	if v.Type().Implements(unmarshalerType) {
		u = v.Interface().(Unmarshaler)
	}
	if v.CanAddr() {
		v = v.Addr()
		if v.Type().Implements(marshalerType) {
			m = v.Interface().(Marshaler)
		}
		if v.Type().Implements(unmarshalerType) {
			u = v.Interface().(Unmarshaler)
		}
	}
	if m == nil || u == nil {
		return nil, false
	}
	return BasicEncoding{encodeMarshaler, decodeMarshaler}.ToEncoding(v), true
}

func matchTextMarshaler(v reflect.Value) (*Encoding, bool) {
	var m encoding.TextMarshaler
	var u encoding.TextUnmarshaler
	if v.Type().Implements(textMarshalerType) {
		m = v.Interface().(encoding.TextMarshaler)
	}
	if v.Type().Implements(textUnmarshalerType) {
		u = v.Interface().(encoding.TextUnmarshaler)
	}
	if v.CanAddr() {
		v = v.Addr()
		if v.Type().Implements(textMarshalerType) {
			m = v.Interface().(encoding.TextMarshaler)
		}
		if v.Type().Implements(textUnmarshalerType) {
			u = v.Interface().(encoding.TextUnmarshaler)
		}
	}
	if m == nil || u == nil {
		return nil, false
	}
	return BasicEncoding{encodeTextMarshaler, decodeTextMarshaler}.ToEncoding(v), true
}

func matchArray(v reflect.Value) (*Encoding, bool) {
	if v.Kind() != reflect.Array {
		return nil, false
	}
	return &Encoding{encodeArray(v), decodeArray(v)}, true
}

func encodeArray(v reflect.Value) EncodeFunc {
	return func(c Composer) error {
		return c.ComposeList(v.Len(), func(i int) error {
			return c.ComposeAny(v.Index(i))
		})
	}
}

func decodeArray(v reflect.Value) DecodeFunc {
	return func(parser Parser) error {
		return parser.ParseList(func(i int) error {
			elem := reflect.Value{}
			if i < v.Len() {
				elem = v.Index(i)
			}
			if err := parser.ParseAny(elem); err != nil {
				return err
			}
			return nil
		})
	}
}

func matchSlice(v reflect.Value) (*Encoding, bool) {
	if v.Kind() != reflect.Slice {
		return nil, false
	}
	return &Encoding{encodeSlice(v), decodeSlice(v)}, true
}

func encodeSlice(v reflect.Value) EncodeFunc {
	return func(c Composer) error {
		if v.IsNil() {
			return composeNil(c)
		}
		return encodeArray(v)(c)
	}
}

func decodeSlice(v reflect.Value) DecodeFunc {
	return func(parser Parser) error {
		if isNil(parser) {
			v.Set(reflect.Zero(v.Type()))
			return nil
		}
		return parser.ParseList(func(i int) error {
			if i == v.Len() {
				v.Set(reflect.Append(v, reflect.New(v.Type().Elem()).Elem()))
			}
			elem := v.Index(i)
			if err := parser.ParseAny(elem); err != nil {
				return err
			}
			return nil
		})
	}
}

func matchStruct(v reflect.Value) (*Encoding, bool) {
	if v.Kind() != reflect.Struct {
		return nil, false
	}
	return &Encoding{encodeStruct(v), decodeStruct(v)}, true
}

func encodeStruct(v reflect.Value) EncodeFunc {
	return func(c Composer) error {
		t := v.Type()
		return c.ComposeList(t.NumField(), func(i int) error {
			if c.Inline() {
				return c.ComposeAny(v.Field(i))
			} else {
				composeValue(c, t.Field(i).Name)
				return c.ComposeList(1, func(int) error {
					return c.ComposeAny(v.Field(i))
				})
			}
		})
	}
}

func decodeStruct(v reflect.Value) DecodeFunc {
	return func(parser Parser) error {
		return parser.ParseList(func(int) error {
			var fieldName string
			if err := parser.ParseAny(reflectValue(&fieldName)); err != nil {
				return err
			}
			elem := reflect.Value{}
			if field := v.FieldByName(string(fieldName)); field.CanSet() {
				elem = field
			}
			if err := parser.ParseAny(elem); err != nil {
				return err
			}
			return nil
		})
	}
}

func matchMap(v reflect.Value) (*Encoding, bool) {
	if v.Kind() != reflect.Map {
		return nil, false
	}
	return &Encoding{encodeMap(v), decodeMap(v)}, true
}

func encodeMap(v reflect.Value) EncodeFunc {
	return func(c Composer) error {
		if v.IsNil() {
			composeNil(c)
			return nil
		}
		var keys stringValues = v.MapKeys()
		sort.Sort(keys)
		return c.ComposeList(v.Len(), func(i int) error {
			composeValue(c, encodeKey(keys[i]))
			return c.ComposeList(1, func(int) error {
				return c.ComposeAny(v.MapIndex(keys[i]))
			})
		})
	}
}

func decodeMap(v reflect.Value) DecodeFunc {
	return func(parser Parser) error {
		if isNil(parser) {
			v.Set(reflect.Zero(v.Type()))
			return nil
		}
		if v.IsNil() {
			v.Set(reflect.MakeMap(v.Type()))
		}
		return parser.ParseList(func(int) error {
			key := reflect.New(v.Type().Key()).Elem()
			if err := parser.ParseAny(key); err != nil {
				return err
			}
			elem := reflect.New(v.Type().Elem()).Elem()
			/*
				if err := parser.GoToOnlyChild(); err != nil {
					return err
				}
				if err := skipColon(parser); err != nil {
					return err
				}
			*/
			if err := parser.ParseAny(elem); err != nil {
				return err
			}
			v.SetMapIndex(key, elem)
			return nil
		})
	}
}

func isNil(parser Parser) bool {
	if val, err := parser.Value(); err == nil {
		return string(val) == "nil"
	}
	return false
}

func skipColon(parser Parser) error {
	if val, err := parser.Value(); err == nil && string(val) == ":" {
		return parser.GoToOnlyChild()
	}
	return nil
}

func encodeKey(v reflect.Value) string {
	buf, _ := MarshalIndent(v, "", "")
	return string(buf)
}

func composeNil(c Composer) error {
	return composeValue(c, "nil")
}

func composeValue(c Composer, s string) error {
	_, err := c.Write([]byte(s))
	return err
}

type stringValues []reflect.Value

func (sv stringValues) Len() int           { return len(sv) }
func (sv stringValues) Swap(i, j int)      { sv[i], sv[j] = sv[j], sv[i] }
func (sv stringValues) Less(i, j int) bool { return sv.get(i) < sv.get(j) }
func (sv stringValues) get(i int) string   { return encodeKey(sv[i]) }
