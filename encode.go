// Copyright 2014, Hǎiliàng Wáng. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package flow

import (
	"fmt"
	"io"
	"reflect"
	"strconv"
)

func Marshal(v interface{}) ([]byte, error) {
	enc := NewEncoder(nil)
	if err := enc.marshal(v); err != nil {
		return nil, err
	}
	return enc.Bytes(), nil
}

// MarshalIndent is like Marshal but applies Indent to format the output.
func MarshalIndent(v interface{}, prefix, indent string) ([]byte, error) {
	enc := NewEncoder(nil)
	if err := enc.marshalIndent(v, prefix, indent); err != nil {
		return nil, err
	}
	return enc.Bytes(), nil
}

type Encoder struct {
	w io.Writer
	refDetector
	composer
}

func NewEncoder(w io.Writer) *Encoder {
	enc := &Encoder{
		w:           w,
		refDetector: newRefDetector()}
	enc.Init("", "\t")
	return enc
}

func (enc *Encoder) marshal(v interface{}) error {
	if err := enc.Encode(v); err != nil {
		return err
	}
	return nil
}

func (enc *Encoder) marshalIndent(v interface{}, prefix, indent string) error {
	enc.Init(prefix, indent)
	return enc.marshal(v)
}

func (enc *Encoder) Encode(v interface{}) error {
	enc.start()
	defer func() {
		enc.stop()
	}()
	rv := reflectValue(v)
	if err := enc.checkCycle(rv); err != nil {
		return err
	}
	if err := enc.ComposeAny(rv); err != nil {
		return err
	}
	if enc.w != nil {
		_, err := enc.w.Write(enc.Bytes())
		return err
	}
	return nil
}

func (enc *Encoder) checkCycle(rv reflect.Value) error {
	enc.populate(rv)
	if enc.serial > 1 {
		if rv.Kind() != reflect.Ptr && !rv.CanAddr() {
			return fmt.Errorf("object with cyclic reference must be addressable, %v", rv.Interface())
		}
	}
	return nil
}

// encode never returns an error, it may panics with bytes.ErrTooLarge.
func (enc *Encoder) ComposeAny(v reflect.Value) error {
	if !v.IsValid() {
		enc.encodeNil()
		return nil
	}
	if enc.encodeRefLabel(v) {
		return nil
	}
	if ok, err := enc.encodeAliasedBasicType(v); ok {
		return err
	} else {
		return enc.composeAnyValue(v)
	}
}

func (enc *Encoder) composeAnyValue(v reflect.Value) error {
	if ok, err := enc.encodeCustom(v); ok {
		return err
	}
	switch v.Kind() {
	case reflect.Ptr:
		enc.encodePtr(v)
	case reflect.Interface:
		enc.encodeInterface(v)
	case reflect.Invalid, reflect.Chan, reflect.Func, reflect.UnsafePointer:
		return fmt.Errorf("unsupported variable type: %s", v.Type().String())
	}
	return nil
}

func (enc *Encoder) encodeCustom(v reflect.Value) (bool, error) {
	for _, match := range matchFuncs {
		if encoding, ok := match(v); ok && encoding.Encode != nil {
			return true, encoding.Encode(enc)
		}
	}
	return false, nil
}

func (enc *Encoder) encodeRefLabel(v reflect.Value) bool {
	if v.CanAddr() {
		key := newRefKey(v)
		id := enc.getPtrID(key)
		if id > 0 && !enc.m[key].defined {
			enc.define(key)
			enc.WriteString("^" + strconv.Itoa(id))
			return true
		}
	}
	return false
}

func (enc *Encoder) encodePtr(v reflect.Value) {
	if v.IsNil() {
		enc.encodeNil()
		return
	}
	key := newRefKey(v.Elem())
	id := enc.getPtrID(key)
	if id > 0 {
		if enc.m[key].defined {
			enc.WriteString("^" + strconv.Itoa(id))
		} else {
			enc.define(key)
			enc.WriteString("^" + strconv.Itoa(id))
			enc.ComposeList(1, func(int) error {
				return enc.ComposeAny(v.Elem())
			})
		}
	} else {
		enc.ComposeAny(v.Elem())
	}
}

func (enc *Encoder) encodeInterface(v reflect.Value) {
	if v.IsNil() {
		enc.encodeNil()
		return
	}
	v = v.Elem()
	enc.encodeTypeName(v)
	enc.ComposeList(1, func(int) error {
		return enc.ComposeAny(v)
	})
}

func (enc *Encoder) encodeAliasedBasicType(v reflect.Value) (bool, error) {
	switch v.Kind() {
	case reflect.Bool, reflect.Int8, reflect.Int16, reflect.Int32,
		reflect.Int64, reflect.Int, reflect.Uint8, reflect.Uint16,
		reflect.Uint32, reflect.Uint64, reflect.Uint, reflect.Uintptr,
		reflect.Float32, reflect.Float64, reflect.Complex64, reflect.Complex128,
		reflect.String:
		if v.Kind().String() != v.Type().Name() {
			enc.encodeTypeName(v)
			return true, enc.ComposeList(1, func(int) error {
				return enc.composeAnyValue(v)
			})
		}
	}
	return false, nil
}

func (enc *Encoder) encodeTypeName(v reflect.Value) {
	typ := indirectType(v.Type()).Name()
	composeValue(enc, "!"+typ)
}

func indirectType(t reflect.Type) reflect.Type {
	if t.Kind() == reflect.Ptr {
		return indirectType(t.Elem())
	}
	return t
}

type refInfo struct {
	id      int
	defined bool
}

type refKey struct {
	addr uintptr
	typ  reflect.Type
}

func newRefKey(v reflect.Value) refKey {
	return refKey{v.Addr().Pointer(), v.Type()}
}

type refDetector struct {
	m      map[refKey]refInfo
	serial int
}

func newRefDetector() refDetector {
	return refDetector{make(map[refKey]refInfo), 1}
}

func (d *refDetector) getPtrID(key refKey) int {
	if ref := d.m[key]; ref.id > 0 {
		return ref.id
	}
	return 0
}

func (d *refDetector) define(key refKey) {
	ref := d.m[key]
	ref.defined = true
	d.m[key] = ref
}

func (d *refDetector) add(key refKey) {
	ref := d.m[key]
	switch ref.id {
	case 0:
		ref.id = -1
	case -1:
		ref.id = d.serial
		d.serial++
	}
	d.m[key] = ref
}

func (d *refDetector) populate(v reflect.Value) {
	if v.Kind() != reflect.Ptr && v.CanAddr() {
		key := newRefKey(v)
		d.add(key)
		if d.m[key].id > 0 {
			return
		}
	}
	switch v.Kind() {
	case reflect.Ptr:
		d.populate(v.Elem())
	case reflect.Array, reflect.Slice:
		for i := 0; i < v.Len(); i++ {
			d.populate(v.Index(i))
		}
	case reflect.Struct:
		for i := 0; i < v.Type().NumField(); i++ {
			d.populate(v.Field(i))
		}
	case reflect.Map:
		for _, k := range v.MapKeys() {
			d.populate(v.MapIndex(k))
		}
	}
}

func reflectValue(v interface{}) reflect.Value {
	rv, ok := v.(reflect.Value)
	if ok {
		return rv
	}
	return reflect.ValueOf(v)
}
