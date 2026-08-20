package domain

type Node struct {
	Path      string
	Value     any
	Sensitive bool
	Children  []Node
}
type Explanation struct {
	Decision  string
	Reason    string
	Version   string
	Root      Node
	Truncated bool
}

func (n Node) Clone() Node {
	return n
}

func (e Explanation) Clone() Explanation {
	e.Root = e.Root.Clone()
	return e
}
