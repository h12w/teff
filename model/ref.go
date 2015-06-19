package model

import (
	"reflect"
	"strconv"
)

// maker makes a new List
type maker struct {
	m      map[uintptr]*Node
	serial int
}
type nodeCounter struct {
	n      *Node
	cnt    int
	serial int
}

// filler fills from a list
type filler struct {
	m map[RefID]reflect.Value
}

func newFiller() *filler {
	return &filler{make(map[RefID]reflect.Value)}
}

func newMaker() *maker {
	return &maker{
		m:      make(map[uintptr]*Node),
		serial: 1,
	}
}

func (m *maker) register(p uintptr, node *Node) {
	m.m[p] = node
}

func (f *filler) register(refID RefID, v reflect.Value) {
	f.m[refID] = v
}

func (m *maker) find(addr uintptr) (*Node, bool) {
	if node, ok := m.m[addr]; ok {
		if node.RefID == "" {
			node.RefID = RefID(strconv.Itoa(m.serial))
			m.serial++
		}
		return node, true
	}
	return nil, false
}

func (f *filler) value(refID RefID) reflect.Value {
	return f.m[refID]
}

/*
func (m *maker) populate(v reflect.Value) *maker {
	s := []reflect.Value{v}
	for len(s) > 0 {
		s, v = s[:len(s)-1], s[len(s)-1]
		for _, addr := range addresses(v) {
			m.register(addr, &Node{})
		}
		switch v.Type().Kind() {
		case reflect.Slice, reflect.Array:
			for i := 0; i < v.Len(); i++ {
				s = append(s, v.Index(i))
			}
		case reflect.Struct:
			for i := 0; i < v.NumField(); i++ {
				s = append(s, v.Field(i))
			}
		case reflect.Map:
		}
	}
	return m
}
*/

func (m *maker) nodeFromPtr(v reflect.Value) (*Node, error) {
	n := &Node{}
	if v.IsNil() {
		return n, nil // avoid infinite loop
	}
	for v.Type().Kind() == reflect.Ptr {
		if refNode, ok := m.find(v.Pointer()); ok {
			n.Value = refNode.RefID
			return n, nil
		}
		v = reflect.Indirect(v)
	}
	return m.node(leaf(v))
}

func (f *filler) ptrFromNode(n *Node, v reflect.Value) error {
	if refID, ok := n.Value.(RefID); ok {
		ref := f.value(refID)
		if ref.Type() != v.Type() {
			ref = ref.Addr()
			for v.Type() != ref.Type() {
				v = allocIndirect(v)
			}
		}
		v.Set(ref)
		return nil
	}
	return f.fromNode(n, allocLeaf(v))
}

func leaf(v reflect.Value) reflect.Value {
	for v.Type().Kind() == reflect.Ptr && !v.IsNil() {
		v = reflect.Indirect(v)
	}
	return v
}

func addresses(v reflect.Value) (addrs []uintptr) {
	if v.CanAddr() {
		v = v.Addr()
	}
	for v.Type().Kind() == reflect.Ptr && !v.IsNil() {
		addrs = append(addrs, v.Pointer())
		v = reflect.Indirect(v)
	}
	return
}

func allocLeaf(v reflect.Value) reflect.Value {
	for v.Type().Kind() == reflect.Ptr {
		v = allocIndirect(v)
	}
	return v
}

func allocIndirect(v reflect.Value) reflect.Value {
	if v.IsNil() {
		v.Set(reflect.New(v.Type().Elem()))
	}
	return reflect.Indirect(v)
}
