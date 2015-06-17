package core

type (
	Node struct {
		Value       string
		IsReference bool
		List        List
		Annotations []string
	}
	List []Node
)
