package model

import (
	"fmt"
	"reflect"
)

func New(v interface{}) (*Node, error) {
	if v == nil {
		return nil, nil
	}
	return newMaker().toNode(reflect.ValueOf(v))
}

func (n *Node) Fill(v interface{}) error {
	if v == nil {
		return nil
	}
	return newFiller().nodeTo(n, reflect.ValueOf(v))
}

func (m *maker) toNode(v reflect.Value) (*Node, error) {
	var c C
	var err error
	switch v.Type().Kind() {
	case reflect.Int, reflect.String:
		c, err = m.toValue(v)
	case reflect.Slice, reflect.Array:
		c, err = m.toArray(v)
	case reflect.Ptr:
		return m.ptrToNode(v)
	default:
		err = fmt.Errorf("maker.toNode: unsupported type: %v", v.Type())
	}
	if err != nil {
		return nil, err
	}
	return &Node{C: c}, nil
}

func (f *filler) nodeTo(node *Node, v reflect.Value) error {
	switch v.Type().Kind() {
	case reflect.Int, reflect.String:
		if value, ok := node.C.(Value); ok {
			return f.valueTo(value, v)
		}
	case reflect.Slice, reflect.Array:
		if array, ok := node.C.(Array); ok {
			return f.arrayTo(array, v)
		}
	case reflect.Ptr:
		return f.nodeToPtr(node, v)
	}
	return fmt.Errorf("filler.nodeTo: unsupported type: %v", v.Type())
}

func (m *maker) toArray(v reflect.Value) (Array, error) {
	a := make(Array, v.Len())
	for i := 0; i < v.Len(); i++ {
		node, err := m.toNode(v.Index(i))
		if err != nil {
			return nil, err
		}
		a[i] = node
	}
	return a, nil
}

func (f *filler) arrayTo(a Array, v reflect.Value) error {
	for i, n := range a {
		v.Set(reflect.Append(v, reflect.New(v.Type().Elem()).Elem()))
		elem := v.Index(i)
		if err := f.nodeTo(n, elem); err != nil {
			return err
		}
	}
	return nil
}

func (m *maker) toValue(v reflect.Value) (Value, error) {
	switch v.Type().Kind() {
	case reflect.Int:
		return Value{int(v.Int())}, nil
	case reflect.String:
		return Value{v.String()}, nil
	}
	return Value{}, fmt.Errorf("maker.toValue: unsupported type: %v", v.Type())
}

func (f *filler) valueTo(value Value, v reflect.Value) error {
	switch v.Type().Kind() {
	case reflect.Int, reflect.String:
		v.Set(reflect.ValueOf(value.V))
		return nil
	case reflect.Ptr:
		return f.valueToPtr(value, v)
	}
	return fmt.Errorf("filler.valueTo: unsupported type: %v", v.Type())
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
