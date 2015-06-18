/*
Package model converts almost any Go data structures into a tree model:
1. references are replaced with RefIDs
2. pointer & interfaces are replaced with values
3. reflections are hidden from outside if possible
*/
package model

import (
	"errors"
	"reflect"
	"strconv"
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
	if v.Type().Kind() == reflect.Ptr {
		v = indirect(v)
	}
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
	}
	return nil, errors.New("maker.list: unsupported type")
}

func (f *filler) fromList(l List, v reflect.Value) error {
	if v.Type().Kind() == reflect.Ptr {
		v = allocIndirect(v)
	}
	switch v.Type().Kind() {
	case reflect.Int, reflect.String:
		if len(l) > 0 {
			return f.fromNode(l[0], v)
		}
	case reflect.Slice:
		return f.sliceFromList(l, v)
	case reflect.Struct:
		return f.structFromList(l, v)
	}
	return errors.New("List.fill: unsupported type")
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
		if fieldName, ok := n.Value.(Identifier); ok {
			if field := v.FieldByName(string(fieldName)); field.IsValid() {
				if len(n.List) > 0 {
					if err := f.fromNode(n.List[0], field); err != nil {
						return err
					}
				}
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
		if v.IsNil() {
			return &Node{}, nil // avoid infinite loop
		}
		found := false
		for {
			addr := v.Pointer()
			if refNode, ok := m.find(addr); ok {
				n = &Node{Value: refNode.RefID}
				found = true
				break
			}

			if !v.IsNil() {
				v = reflect.Indirect(v)
			}
			if v.Type().Kind() != reflect.Ptr || v.IsNil() {
				break
			}
		}
		if !found {
			n, err = m.node(indirect(v))
		}
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
		if refID, ok := n.Reference(); ok {
			if ref := f.value(refID); ref.IsValid() {
				if ref.Type() == v.Type() {
					v.Set(ref)
				} else {
					if reflect.PtrTo(ref.Type()) == v.Type() {
						v.Set(ref.Addr())
					} else {
						//for i := 0; i < 10; i++ {
						//	if ref.CanAddr() {
						//		ref = ref.Addr()
						//	} else {
						//		nref := reflect.New(reflect.PtrTo(ref.Type()))
						//		nref.Set(ref)
						//		ref = nref
						//	}
						//	if ref.Type() == v.Type() {
						//		v.Set(ref)
						//	}
						//}
					}
				}
			}
		} else {
			err = f.fromNode(n, allocIndirect(v))
		}
	default:
		err = errors.New("Node.fill: unsupported type")
	}
	if err == nil && n.RefID != "" {
		f.register(n.RefID, v)
	}
	return
}

// maker makes a new List
type maker struct {
	m      map[uintptr]*Node
	serial int
}

// filler fills from a list
type filler struct {
	m map[RefID]reflect.Value
}

func newFiller() *filler {
	return &filler{make(map[RefID]reflect.Value)}
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

func (f *filler) register(refID RefID, v reflect.Value) {
	f.m[refID] = v
}

func (m *maker) find(addr uintptr) (*Node, bool) {
	if node, ok := m.m[addr]; ok {
		if node.RefID == "" {
			node.RefID = RefID(strconv.Itoa(m.serial))
			m.serial++
		}
		return node, true
	}
	return nil, false
}

func (f *filler) value(refID RefID) reflect.Value {
	return f.m[refID]
}

func indirect(v reflect.Value) reflect.Value {
	for v.Type().Kind() == reflect.Ptr && !v.IsNil() {
		v = reflect.Indirect(v)
	}
	return v
}

func addresses(v reflect.Value) (addrs []uintptr) {
	if v.CanAddr() {
		v = v.Addr()
	}
	for v.Type().Kind() == reflect.Ptr && !v.IsNil() {
		addrs = append(addrs, v.Pointer())
		v = reflect.Indirect(v)
	}
	return
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
