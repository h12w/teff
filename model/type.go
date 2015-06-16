package model

type (
	List []*Node
	Node struct {
		RefID RefID
		Value interface{}
		List  List
	}
	RefID      string
	IdentValue struct {
		Ident Identifier
		Value interface{}
	}
	Identifier string
)

func (n *Node) reference() (label RefID, ok bool) {
	v := n.Value
	if iv, ok := n.Value.(IdentValue); ok {
		v = iv.Value
	}
	id, ok := v.(RefID)
	return id, ok
}
