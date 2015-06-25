package model

import (
	"fmt"
	"reflect"
	"strings"
)

const (
	ValueNodeType NodeType = iota
	ArrayNodeType
	MapNodeType
)

type (
	NodeType int
	Node     interface {
		Type() NodeType
		RefID() RefID
		SetRefID(RefID)
		String() string
	}
	Value struct {
		V interface{}
		nodeBase
	}
	Map struct {
		KV []KeyValue
		nodeBase
	}
	KeyValue struct {
		K interface{}
		V Node
	}
	Array struct {
		L []Node
		nodeBase
	}
	nodeBase struct {
		r RefID
	}
	RefID string
)

func (b *nodeBase) RefID() RefID {
	return b.r
}
func (b *nodeBase) SetRefID(refID RefID) {
	b.r = refID
}

func (n *Value) Type() NodeType {
	return ValueNodeType
}
func (n *Map) Type() NodeType {
	return MapNodeType
}
func (n *Array) Type() NodeType {
	return ArrayNodeType
}

func (n *Value) String() string {
	return fmt.Sprintf("%v%s(%v)", n.V, n.nodeBase.String(), reflect.TypeOf(n.V).Name())
}

func (n *Array) String() string {
	ss := make([]string, len(n.L))
	for i := range ss {
		ss[i] = n.L[i].String()
	}
	return "{" + strings.Join(ss, ", ") + "}"
}

func (n *Map) String() string {
	ss := make([]string, len(n.KV))
	for i := range ss {
		ss[i] = fmt.Sprint(n.KV[i].K) + ":" + n.KV[i].V.String()
	}
	return "{" + strings.Join(ss, ", ") + "}"
}

func (n *nodeBase) String() string {
	r := string(n.r)
	if r != "" {
		r = "^" + r
	}
	return r
}
