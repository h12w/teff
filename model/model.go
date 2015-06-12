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

type List []*Node

type Node struct {
	Label Label
	Value interface{}
	List  List
}
type Label string

func New(v interface{}) (List, error) {
	if v == nil {
		return nil, nil
	}
	return newMaker().newList(reflect.ValueOf(v))
}

func (l List) Fill(v interface{}) error {
	if v == nil {
		return nil
	}
	return l.fill(reflect.ValueOf(v))
}

// maker makes a new List
type maker struct {
	m      map[uintptr]*Node
	serial int
}

func newMaker() *maker {
	return &maker{
		m:      make(map[uintptr]*Node),
		serial: 1,
	}
}

func (m *maker) label(addr uintptr) (Label, bool) {
	if node, ok := m.m[addr]; ok {
		if node.Label == "" {
			node.Label = Label(strconv.Itoa(m.serial))
			p(node.Label)
			m.serial++
		}
		return node.Label, true
	}
	return Label(0), false
}

func (m *maker) register(p uintptr, node *Node) {
	m.m[p] = node
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
	case reflect.Ptr:
		return m.newList(indirect(v))
	}
	return nil, errors.New("newList: unsupported type")
}

func (m *maker) newNode(v reflect.Value) (*Node, error) {
	switch v.Type().Kind() {
	case reflect.Int, reflect.String:
		return &Node{Value: v.Interface()}, nil
	case reflect.Slice:
		list, err := m.newList(v)
		if err != nil {
			return nil, err
		}
		return &Node{List: list}, nil
	case reflect.Ptr:
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
	return nil, errors.New("newNode: unsupported type")
}

func (l List) fill(v reflect.Value) error {
	switch v.Type().Kind() {
	case reflect.Int, reflect.String:
		if len(l) > 0 {
			return l[0].fill(v)
		}
	case reflect.Slice:
		for i, n := range l {
			v.Set(reflect.Append(v, reflect.New(v.Type().Elem()).Elem()))
			elem := v.Index(i)
			if err := n.fill(elem); err != nil {
				return err
			}
		}
		return nil
	case reflect.Ptr:
		return l.fill(allocIndirect(v))
	}
	return errors.New("List.fill: unsupported type")
}

func (n Node) fill(v reflect.Value) error {
	switch v.Type().Kind() {
	case reflect.Int, reflect.String:
		if _, ok := n.Value.(Label); ok {
			return nil
		}
		v.Set(reflect.ValueOf(n.Value))
		return nil
	case reflect.Slice:
		return n.List.fill(v)
	case reflect.Ptr:
		return n.fill(allocIndirect(v))
	}
	return errors.New("Node.fill: unsupported type")
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
