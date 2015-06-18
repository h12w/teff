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
