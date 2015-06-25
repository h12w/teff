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
		m:      make(map[uintptr]*Node),
		serial: 1,
	}
}

func (m *maker) register(v reflect.Value, node *Node) {
	if node == nil {
		return
	}
	for _, p := range addresses(v) {
		m.m[p] = node
	}
}

func (f *filler) register(refID RefID, v reflect.Value) {
	if refID != "" {
		f.m[refID] = v
	}
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

func (m *maker) ptrToNode(v reflect.Value, node *Node) error {
	if v.IsNil() {
		return nil // avoid infinite loop
	}
	for v.Type().Kind() == reflect.Ptr {
		if refNode, ok := m.find(v.Pointer()); ok {
			node.Value = refNode.RefID
			return nil
		}
		v = reflect.Indirect(v)
	}
	return m.objectToNode(v, node)
}

func (f *filler) nodeToPtr(n *Node, v reflect.Value) error {
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
	return f.nodeToObject(n, allocIndirect(v))
}

func (m *maker) ptrToList(v reflect.Value) (List, error) {
	return m.objectToList(reflect.Indirect(v))
}

func (f *filler) listToPtr(l List, v reflect.Value) error {
	return f.listToObject(l, allocIndirect(v))
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

func allocIndirect(v reflect.Value) reflect.Value {
	alloc(v)
	return reflect.Indirect(v)
}

func alloc(v reflect.Value) reflect.Value {
	if v.IsNil() {
		v.Set(reflect.New(v.Type().Elem()))
	}
	return v
}
