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

// Reference returns RefID, true if the node is a reference to another node.
// Otherwise, it will return 0, false.
func (n *Node) Reference() (refID RefID, ok bool) {
	v := n.Value
	if iv, ok := n.Value.(IdentValue); ok {
		v = iv.Value
	}
	id, ok := v.(RefID)
	return id, ok
}
