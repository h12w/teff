package model

import (
	"reflect"
	"strconv"
)

// maker makes a new List
type maker struct {
	m      map[uintptr]nodeRegistry
	serial int
}
type nodeRegistry struct {
	node     *Node
	isSource bool
}

// filler fills from a list
// TODO: fill lazily
type filler struct {
	m map[RefID]reflect.Value
}

func newFiller() *filler {
	return &filler{make(map[RefID]reflect.Value)}
}

func newMaker() *maker {
	return &maker{
		m:      make(map[uintptr]nodeRegistry),
		serial: 1,
	}
}

func (m *maker) find(addr uintptr) (*Node, bool) {
	if r, ok := m.m[addr]; ok {
		if r.node.RefID == "" {
			r.node.RefID = RefID(strconv.Itoa(m.serial))
			m.serial++
		}
		return r.node, true
	}
	return nil, false
}

func (f *filler) value(refID RefID) reflect.Value {
	return f.m[refID]
}

func (f *filler) nodeToPtr(n *Node, v reflect.Value) error {
	if value, ok := n.C.(Value); ok {
		return f.valueToPtr(value, v)
	}
	return f.nodeTo(n, allocIndirect(v))
}

func (m *maker) ptrToNode(v reflect.Value) (*Node, error) {
	if v.IsNil() {
		return nil, nil // avoid infinite loop
	}
	for v.Type().Kind() == reflect.Ptr {
		if refNode, ok := m.find(v.Pointer()); ok {
			return &Node{C: Value{refNode.RefID}}, nil
		}
		v = reflect.Indirect(v)
	}
	return m.toNode(v)
}

func (f *filler) valueToPtr(v Value, o reflect.Value) error {
	if refID, ok := v.V.(RefID); ok {
		ref := f.value(refID)
		if ref.Type() != o.Type() {
			ref = ref.Addr()
			for o.Type() != ref.Type() {
				o = allocIndirect(o)
			}
		}
		o.Set(ref)
		return nil
	}
	return f.valueTo(v, allocIndirect(o))
}
