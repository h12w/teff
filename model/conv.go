package model

import (
	"fmt"
	"reflect"
)

func New(v interface{}) (Node, error) {
	if v == nil {
		return nil, nil
	}
	return newMaker().objectToNode(reflect.ValueOf(v))
}

func Fill(n Node, v interface{}) error {
	if v == nil {
		return nil
	}
	return newFiller().nodeToObject(n, reflect.ValueOf(v))
}

func (m *maker) objectToNode(v reflect.Value) (Node, error) {
	switch v.Type().Kind() {
	case reflect.Int, reflect.String:
		return m.objectToValue(v)
	case reflect.Slice, reflect.Array:
		return m.arrayToArray(v)
	case reflect.Ptr:
		return m.ptrToNode(v)
	}
	return nil, fmt.Errorf("maker.objectToNode: unsupported type: %v", v.Type())
}

func (f *filler) nodeToObject(node Node, v reflect.Value) error {
	switch v.Type().Kind() {
	case reflect.Int, reflect.String:
		if value, ok := node.(*Value); ok {
			return f.valueToObject(value, v)
		}
	case reflect.Slice, reflect.Array:
		if array, ok := node.(*Array); ok {
			return f.arrayToArray(array, v)
		}
	case reflect.Ptr:
		return f.nodeToPtr(node, v)
	}
	return fmt.Errorf("filler.nodeToObject: unsupported type: %v", v.Type())
}

func (m *maker) arrayToArray(v reflect.Value) (*Array, error) {
	nodes := make([]Node, v.Len())
	for i := 0; i < v.Len(); i++ {
		node, err := m.objectToNode(v.Index(i))
		if err != nil {
			return nil, err
		}
		nodes[i] = node
	}
	return &Array{L: nodes}, nil
}

func (f *filler) arrayToArray(a *Array, v reflect.Value) error {
	for i, n := range a.L {
		v.Set(reflect.Append(v, reflect.New(v.Type().Elem()).Elem()))
		elem := v.Index(i)
		if err := f.nodeToObject(n, elem); err != nil {
			return err
		}
	}
	return nil
}

func (m *maker) objectToValue(v reflect.Value) (*Value, error) {
	switch v.Type().Kind() {
	case reflect.Int:
		return &Value{V: int(v.Int())}, nil
	case reflect.String:
		return &Value{V: v.String()}, nil
	}
	return nil, fmt.Errorf("maker.objectToValue: unsupported type: %v", v.Type())
}

func (f *filler) valueToObject(value *Value, v reflect.Value) error {
	switch v.Type().Kind() {
	case reflect.Int, reflect.String:
		v.Set(reflect.ValueOf(value.V))
		return nil
	case reflect.Ptr:
		return f.valueToPtr(value, v)
	}
	return fmt.Errorf("filler.valueToObject: unsupported type: %v", v.Type())
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
