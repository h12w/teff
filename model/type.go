package model

type (
	List []*Node
	Node struct {
		RefID RefID
		Value interface{}
		List  List
	}
	RefID      string
	Identifier string
)

// Reference returns RefID, true if the node is a reference to another node.
// Otherwise, it will return "", false.
func (n *Node) Reference() (refID RefID, ok bool) {
	id, ok := n.Value.(RefID)
	return id, ok
}
