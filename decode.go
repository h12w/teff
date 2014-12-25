// Copyright 2014, Hǎiliàng Wáng. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package flow

import (
	"fmt"
	"io"
	"reflect"
)

type Decoder struct {
	*parser
	refSetter
}

func NewDecoder(r io.Reader) *Decoder {
	return &Decoder{
		newParser(r),
		newRefSetter(),
	}
}

func (dec *Decoder) Decode(v interface{}) error {
	rv, ok := v.(reflect.Value)
	if !ok {
		rv = reflect.ValueOf(v)
	}
	if err := dec.next(); err != nil {
		return err
	}
	if err := dec.ParseAny(rv); err != nil {
		return err
	}
	dec.setAllRef()
	return nil
}

func (dec *Decoder) ParseAny(v reflect.Value) (err error) {
	if !v.CanSet() && v.Kind() != reflect.Ptr { // TODO: interface should also be allowed.
		return fmt.Errorf("unsetable nonpointer value: %v", v)
	}
	if v.Kind() == reflect.Ptr {
		v = alloc(v).Elem()
	}
	for dec.isRef() || dec.isType() {
		if dec.isRef() {
			id := string(dec.Token().Value[1:])
			if err := dec.next(); err != nil {
				return err
			}
			if dec.isSepOrListEnd() {
				dec.addDstRef(id, v)
				return nil
			} else {
				dec.addSrcRef(id, v)
			}
		} else if dec.isType() {
			typ := string(dec.Token().Value[1:])
			t, ok := nameToType[typ]
			if !ok {
				return fmt.Errorf("type %s is not registered.", typ)
			}
			nv := reflect.New(t).Elem()
			ov := v
			v = nv
			defer func() {
				ov.Set(nv)
			}()
			if err := dec.next(); err != nil {
				return err
			}
		}
	}

	defer func() {
		if e := dec.next(); e != nil && e != io.EOF {
			err = e
		}
	}()
	for _, match := range matchFuncs {
		if encoding, ok := match(v); ok && encoding.Decode != nil {
			return encoding.Decode(dec)
		}
	}
	return fmt.Errorf("no decoding method defined for type: %v", v.Type())
}

func alloc(v reflect.Value) reflect.Value {
	if v.IsNil() {
		v.Set(reflect.New(v.Type().Elem()))
	}
	if ve := v.Elem(); ve.Kind() == reflect.Ptr {
		return alloc(ve)
	}
	return v
}

type refValue struct {
	src reflect.Value
	dst []reflect.Value
}

type refSetter struct {
	m map[string]refValue
}

func newRefSetter() refSetter {
	return refSetter{make(map[string]refValue)}
}

func (s *refSetter) setAllRef() {
	for _, refVal := range s.m {
		for _, dst := range refVal.dst {
			dst.Set(refVal.src)
		}
	}
}

func (s *refSetter) addSrcRef(id string, v reflect.Value) {
	refVal := s.m[id]
	refVal.src = v
	s.m[id] = refVal
}

func (s *refSetter) addDstRef(id string, v reflect.Value) {
	refVal := s.m[id]
	refVal.dst = append(refVal.dst, v)
	s.m[id] = refVal
}
