/*
Package model has the following goals:
1. cyclic references are replaced with labels
2. pointer & interfaces are replaced with values
3. reflections are hidden from outside if possible
*/
package model

type List []Node

type Node struct {
	Label string
	Value interface{}
	List  List
}

func New(v interface{}) (List, error) {
	return nil, nil
}

func (l List) Fill(v interface{}) error {
	return nil
}
