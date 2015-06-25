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
	return newMaker().objectToList(reflect.ValueOf(v))
}

func (l List) Fill(v interface{}) error {
	if v == nil {
		return nil
	}
	return newFiller().listToObject(l, reflect.ValueOf(v))
}

func (m *maker) objectToList(v reflect.Value) (List, error) {
	switch v.Type().Kind() {
	case reflect.Int, reflect.String:
		var node Node
		err := m.objectToNode(v, &node)
		if err != nil {
			return nil, err
		}
		return List{&node}, nil
	case reflect.Slice:
		return m.sliceToList(v)
	case reflect.Struct:
		return m.structToList(v)
	case reflect.Ptr:
		return m.ptrToList(v)
	}
	return nil, errors.New("maker.list: unsupported type")
}

func (f *filler) listToObject(l List, v reflect.Value) error {
	switch v.Type().Kind() {
	case reflect.Int, reflect.String:
		if len(l) > 0 {
			return f.nodeToObject(l[0], v)
		}
	case reflect.Slice:
		return f.listToSlice(l, v)
	case reflect.Struct:
		return f.listToStruct(l, v)
	case reflect.Ptr:
		return f.listToPtr(l, v)
	}
	return fmt.Errorf("List.fill: unsupported type: %v", v.Type())
}

func (m *maker) sliceToList(v reflect.Value) (List, error) {
	l := make(List, v.Len())
	for i := 0; i < v.Len(); i++ {
		var node Node
		err := m.objectToNode(v.Index(i), &node)
		if err != nil {
			return nil, err
		}
		l[i] = &node
	}
	return l, nil
}

func (f *filler) listToSlice(l List, v reflect.Value) error {
	for i, n := range l {
		v.Set(reflect.Append(v, reflect.New(v.Type().Elem()).Elem()))
		elem := v.Index(i)
		if err := f.nodeToObject(n, elem); err != nil {
			return err
		}
	}
	return nil
}

func (m *maker) structToList(v reflect.Value) (List, error) {
	t := v.Type()
	l := make(List, v.NumField())
	for i := 0; i < v.NumField(); i++ {
		var node Node
		err := m.objectToNode(v.Field(i), &node)
		if err != nil {
			return nil, err
		}
		l[i] = &Node{Value: Identifier(t.Field(i).Name), List: List{&node}}
	}
	return l, nil
}

func (f *filler) listToStruct(l List, v reflect.Value) error {
	for _, n := range l {
		if len(n.List) == 0 {
			continue
		}
		if fieldName, ok := n.Value.(Identifier); ok {
			field := v.FieldByName(string(fieldName))
			if err := f.nodeToObject(n.List[0], field); err != nil {
				return err
			}
		}
	}
	return nil
}

func (m *maker) objectToNode(v reflect.Value, n *Node) (err error) {
	switch v.Type().Kind() {
	case reflect.Int:
		n.Value = int(v.Int())
	case reflect.String:
		n.Value = v.String()
	case reflect.Slice:
		n.List, err = m.objectToList(v)
	case reflect.Ptr:
		err = m.ptrToNode(v, n)
	default:
		err = errors.New("node: unsupported type")
	}
	m.register(v, n)
	return
}

func (f *filler) nodeToObject(n *Node, v reflect.Value) (err error) {
	f.register(n.RefID, v)
	switch v.Type().Kind() {
	case reflect.Int, reflect.String:
		v.Set(reflect.ValueOf(n.Value))
	case reflect.Slice:
		err = f.listToObject(n.List, v)
	case reflect.Ptr:
		err = f.nodeToPtr(n, v)
	default:
		err = errors.New("Node.fill: unsupported type")
	}
	return
}
