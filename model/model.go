/*
Package model converts almost any Go data structures into a tree model:
1. cyclic references are replaced with labels
2. pointer & interfaces are replaced with values
3. reflections are hidden from outside if possible
*/
package model

import (
	"errors"
	"reflect"
	"strconv"
)

type (
	List []*Node
	Node struct {
		Label Label
		Value interface{}
		List  List
	}
	Label      string
	IdentValue struct {
		Ident Identifier
		Value interface{}
	}
	Identifier string
)

func New(v interface{}) (List, error) {
	if v == nil {
		return nil, nil
	}
	return newMaker().newList(reflect.ValueOf(v))
}

func Fill(l List, v interface{}) error {
	if v == nil {
		return nil
	}
	return newFiller().fillList(l, reflect.ValueOf(v))
}

func (m *maker) newList(v reflect.Value) (List, error) {
	switch v.Type().Kind() {
	case reflect.Int, reflect.String:
		node, err := m.newNode(v)
		if err != nil {
			return nil, err
		}
		return List{node}, nil
	case reflect.Slice:
		l := make(List, v.Len())
		for i := 0; i < v.Len(); i++ {
			node, err := m.newNode(v.Index(i))
			if err != nil {
				return nil, err
			}
			l[i] = node
		}
		return l, nil
	case reflect.Struct:
		l := make(List, v.NumField())
		for i := 0; i < v.NumField(); i++ {
			node, err := m.newNode(v.Field(i))
			if err != nil {
				return nil, err
			}
			if node.List == nil {
				if label, ok := node.Value.(Label); ok {
					node.Value = label
				} else {
					node.Value = IdentValue{Identifier(v.Type().Field(i).Name), node.Value}
				}
			}
			l[i] = node
		}
		return l, nil
	case reflect.Ptr:
		return m.newList(indirect(v))
	}
	return nil, errors.New("newList: unsupported type")
}

func (f *filler) fillList(l List, v reflect.Value) error {
	switch v.Type().Kind() {
	case reflect.Int, reflect.String:
		if len(l) > 0 {
			return f.fillNode(l[0], v)
		}
	case reflect.Slice:
		for i, n := range l {
			v.Set(reflect.Append(v, reflect.New(v.Type().Elem()).Elem()))
			elem := v.Index(i)
			if err := f.fillNode(n, elem); err != nil {
				return err
			}
		}
		return nil
	case reflect.Struct:
		for i := 0; i < v.NumField(); i++ {
			fieldName := Identifier(v.Type().Field(i).Name)
			for _, n := range l {
				iv := n.Value.(IdentValue)
				if iv.Ident == fieldName {
					v.Field(i).Set(reflect.ValueOf(iv.Value))
					break
				}
			}
		}
		return nil
	case reflect.Ptr:
		return f.fillList(l, allocIndirect(v))
	}
	return errors.New("List.fill: unsupported type")
}

func (m *maker) newNode(v reflect.Value) (*Node, error) {
	switch v.Type().Kind() {
	case reflect.Int:
		return &Node{Value: int(v.Int())}, nil
	case reflect.String:
		return &Node{Value: v.String()}, nil
	case reflect.Slice:
		list, err := m.newList(v)
		if err != nil {
			return nil, err
		}
		return &Node{List: list}, nil
	case reflect.Ptr:
		return m.nodeFromPtr(v)
	}
	return nil, errors.New("newNode: unsupported type")
}

func (f *filler) fillNode(n *Node, v reflect.Value) error {
	switch v.Type().Kind() {
	case reflect.Int, reflect.String:
		if _, ok := n.Value.(Label); ok {
			return nil
		}
		v.Set(reflect.ValueOf(n.Value))
		return nil
	case reflect.Slice:
		return f.fillList(n.List, v)
	case reflect.Ptr:
		return f.fillPtrFromNode(n, v)
	}
	return errors.New("Node.fill: unsupported type")
}

func (m *maker) nodeFromPtr(v reflect.Value) (*Node, error) {
	if v.IsNil() {
		return nil, nil // avoid infinite loop
	}
	addr := v.Pointer()
	if label, ok := m.label(addr); ok {
		return &Node{Value: label}, nil
	}
	node, err := m.newNode(indirect(v))
	if err != nil {
		return nil, err
	}
	m.register(addr, node)
	return node, nil
}

func (f *filler) fillPtrFromNode(n *Node, v reflect.Value) error {
	if n.Label != "" {
		f.register(n.Label, v)
	} else if label, ok := n.Value.(Label); ok {
		v.Set(f.value(label))
		return nil
	}
	return f.fillNode(n, allocIndirect(v))
}

// maker makes a new List
type maker struct {
	m      map[uintptr]*Node
	serial int
}

// filler fills from a list
type filler struct {
	m map[Label]reflect.Value
}

func newFiller() *filler {
	return &filler{make(map[Label]reflect.Value)}
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

func (f *filler) register(label Label, v reflect.Value) {
	f.m[label] = v
}

func (m *maker) label(addr uintptr) (Label, bool) {
	if node, ok := m.m[addr]; ok {
		if node.Label == "" {
			node.Label = Label(strconv.Itoa(m.serial))
			m.serial++
		}
		return node.Label, true
	}
	return Label(0), false
}

func (f *filler) value(label Label) reflect.Value {
	return f.m[label]
}

func indirect(v reflect.Value) reflect.Value {
	for v.Type().Kind() == reflect.Ptr && !v.IsNil() {
		v = reflect.Indirect(v)
	}
	return v
}

func allocIndirect(v reflect.Value) reflect.Value {
	for v.Type().Kind() == reflect.Ptr {
		if v.IsNil() {
			v.Set(reflect.New(v.Type().Elem()))
		}
		v = reflect.Indirect(v)
	}
	return v
}
