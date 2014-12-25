// Copyright 2014, Hǎiliàng Wáng. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package flow

import (
	"fmt"
	"reflect"
	"sync"
)

var (
	registerLock sync.RWMutex
)

var nameToType = map[string]reflect.Type{
	"bool":       reflect.TypeOf(bool(false)),
	"int8":       reflect.TypeOf(int8(0)),
	"int16":      reflect.TypeOf(int16(0)),
	"int32":      reflect.TypeOf(int32(0)),
	"int64":      reflect.TypeOf(int64(0)),
	"int":        reflect.TypeOf(int(0)),
	"uint8":      reflect.TypeOf(uint8(0)),
	"uint16":     reflect.TypeOf(uint16(0)),
	"uint32":     reflect.TypeOf(uint32(0)),
	"uint64":     reflect.TypeOf(uint64(0)),
	"uint":       reflect.TypeOf(uint(0)),
	"uintptr":    reflect.TypeOf(uintptr(0)),
	"float32":    reflect.TypeOf(float32(0)),
	"float64":    reflect.TypeOf(float64(0)),
	"complex64":  reflect.TypeOf(complex64(0)),
	"complex128": reflect.TypeOf(complex128(0)),
	"string":     reflect.TypeOf(string("")),
}

func Register(value interface{}) {
	RegisterName(reflect.TypeOf(value).Name(), value)
}

func RegisterName(name string, value interface{}) {
	if name == "" {
		panic("attempt to register empty name")
	}
	registerLock.Lock()
	defer registerLock.Unlock()
	typ := reflect.TypeOf(value)
	if t, ok := nameToType[name]; ok && t != typ {
		panic(fmt.Sprintf("gob: registering duplicate types for %q: %v != %v", name, t, typ))
	}
	nameToType[name] = typ
}
