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
// Otherwise, it will return "", false.
func (n *Node) Reference() (refID RefID, ok bool) {
	id, ok := n.GetValue().(RefID)
	return id, ok
}

func (n *Node) GetValue() interface{} {
	if iv, ok := n.Value.(IdentValue); ok {
		return iv.Value
	}
	return n.Value
}
