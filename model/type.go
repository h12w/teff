package model

import (
	"fmt"
	"reflect"
	"strings"
)

const (
	ValueNode NodeType = iota
	ArrayNode
	MapNode
)

type (
	Node struct {
		RefID RefID
		C
	}
	RefID string
	C     interface {
		String() string
		Type() NodeType
	}
	NodeType int
	Value    struct {
		V interface{}
	}
	Map      []KeyValue
	KeyValue struct {
		K interface{}
		V *Node
	}
	Array []*Node
)

func (Value) Type() NodeType {
	return ValueNode
}
func (Map) Type() NodeType {
	return MapNode
}
func (Array) Type() NodeType {
	return ArrayNode
}

func (n Value) String() string {
	return fmt.Sprintf("%v(%v)", n.V, reflect.TypeOf(n.V).Name())
}

func (n Array) String() string {
	ss := make([]string, len(n))
	for i := range ss {
		ss[i] = n[i].String()
	}
	return "{" + strings.Join(ss, ", ") + "}"
}

func (n Map) String() string {
	ss := make([]string, len(n))
	for i := range ss {
		ss[i] = fmt.Sprint(n[i].K) + ":" + n[i].V.String()
	}
	return "{" + strings.Join(ss, ", ") + "}"
}

func (n Node) String() string {
	r := string(n.RefID)
	if r != "" {
		r = "^" + r
	}
	return r + " " + n.C.String()
}
