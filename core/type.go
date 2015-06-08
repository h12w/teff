package core

type (
	Node struct {
		Value       string
		List        List
		Annotations []string
	}
	List []Node
)
