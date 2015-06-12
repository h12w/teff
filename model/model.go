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
)

type List []Node

type Node struct {
	Label string
	Value interface{}
	List  List
}

func New(v interface{}) (List, error) {
	if v == nil {
		return nil, nil
	}
	return newList(reflect.ValueOf(v))
}

func (l List) Fill(v interface{}) error {
	if v == nil {
		return nil
	}
	return l.fill(reflect.ValueOf(v))
}

func newList(v reflect.Value) (List, error) {
	switch v.Type().Kind() {
	case reflect.Int:
		node, err := newNode(v)
		if err != nil {
			return nil, err
		}
		return List{node}, nil
	case reflect.Slice:
		l := make(List, v.Len())
		for i := 0; i < v.Len(); i++ {
			node, err := newNode(v.Index(i))
			if err != nil {
				return nil, err
			}
			l[i] = node
		}
		return l, nil
	case reflect.Ptr:
		return newList(indirect(v))
	}
	return nil, errors.New("newList: unsupported type")
}

func newNode(v reflect.Value) (Node, error) {
	switch v.Type().Kind() {
	case reflect.Int:
		return Node{Value: v.Interface()}, nil
	}
	return Node{}, errors.New("newNode: unsupported type")
}

func (l List) fill(v reflect.Value) error {
	switch v.Type().Kind() {
	case reflect.Int:
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
	case reflect.Int:
		v.Set(reflect.ValueOf(n.Value))
		return nil
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
