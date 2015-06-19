/*
Package model converts almost any Go data structures into a tree model:
1. references are replaced with RefIDs
2. pointer & interfaces are replaced with values
3. reflections are hidden from outside if possible
*/
package model

import (
	"errors"
	"fmt"
	"reflect"
)

func New(v interface{}) (List, error) {
	if v == nil {
		return nil, nil
	}
	return newMaker().list(reflect.ValueOf(v))
}

func Fill(l List, v interface{}) error {
	if v == nil {
		return nil
	}
	return newFiller().fromList(l, reflect.ValueOf(v))
}

func (m *maker) list(v reflect.Value) (List, error) {
	switch v.Type().Kind() {
	case reflect.Int, reflect.String:
		node, err := m.node(v)
		if err != nil {
			return nil, err
		}
		return List{node}, nil
	case reflect.Slice:
		return m.listFromSlice(v)
	case reflect.Struct:
		return m.listFromStruct(v)
	case reflect.Ptr:
		return m.listFromPtr(v)
	}
	return nil, errors.New("maker.list: unsupported type")
}

func (f *filler) fromList(l List, v reflect.Value) error {
	switch v.Type().Kind() {
	case reflect.Int, reflect.String:
		if len(l) > 0 {
			return f.fromNode(l[0], v)
		}
	case reflect.Slice:
		return f.sliceFromList(l, v)
	case reflect.Struct:
		return f.structFromList(l, v)
	case reflect.Ptr:
		return f.ptrFromList(l, v)
	}
	return fmt.Errorf("List.fill: unsupported type: %v", v.Type())
}

func (m *maker) listFromSlice(v reflect.Value) (List, error) {
	l := make(List, v.Len())
	for i := 0; i < v.Len(); i++ {
		node, err := m.node(v.Index(i))
		if err != nil {
			return nil, err
		}
		l[i] = node
	}
	return l, nil
}

func (f *filler) sliceFromList(l List, v reflect.Value) error {
	for i, n := range l {
		v.Set(reflect.Append(v, reflect.New(v.Type().Elem()).Elem()))
		elem := v.Index(i)
		if err := f.fromNode(n, elem); err != nil {
			return err
		}
	}
	return nil
}

func (m *maker) listFromStruct(v reflect.Value) (List, error) {
	t := v.Type()
	l := make(List, v.NumField())
	for i := 0; i < v.NumField(); i++ {
		node, err := m.node(v.Field(i))
		if err != nil {
			return nil, err
		}
		l[i] = &Node{Value: Identifier(t.Field(i).Name), List: List{node}}
	}
	return l, nil
}

func (f *filler) structFromList(l List, v reflect.Value) error {
	for _, n := range l {
		if len(n.List) == 0 {
			continue
		}
		if fieldName, ok := n.Value.(Identifier); ok {
			field := v.FieldByName(string(fieldName))
			if err := f.fromNode(n.List[0], field); err != nil {
				return err
			}
		}
	}
	return nil
}

func (m *maker) node(v reflect.Value) (n *Node, err error) {
	switch v.Type().Kind() {
	case reflect.Int:
		n = &Node{Value: int(v.Int())}
	case reflect.String:
		n = &Node{Value: v.String()}
	case reflect.Slice:
		list, err := m.list(v)
		if err != nil {
			return nil, err
		}
		n = &Node{List: list}
	case reflect.Ptr:
		n, err = m.nodeFromPtr(v)
	default:
		err = errors.New("node: unsupported type")
	}
	if n != nil {
		for _, addr := range addresses(v) {
			m.register(addr, n)
		}
	}
	return
}

func (f *filler) fromNode(n *Node, v reflect.Value) (err error) {
	switch v.Type().Kind() {
	case reflect.Int, reflect.String:
		v.Set(reflect.ValueOf(n.Value))
	case reflect.Slice:
		err = f.fromList(n.List, v)
	case reflect.Ptr:
		err = f.ptrFromNode(n, v)
	default:
		err = errors.New("Node.fill: unsupported type")
	}
	if err == nil && n.RefID != "" {
		f.register(n.RefID, v)
	}
	return
}
