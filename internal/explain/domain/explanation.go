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
	children := n.Children
	n.Children = make([]Node, len(children))
	for i := range children {
		n.Children[i] = children[i].Clone()
	}
	switch value := n.Value.(type) {
	case []string:
		n.Value = append([]string(nil), value...)
	case []any:
		n.Value = append([]any(nil), value...)
	case map[string]any:
		copyMap := make(map[string]any, len(value))
		for key, item := range value {
			copyMap[key] = item
		}
		n.Value = copyMap
	}
	return n
}

func (e Explanation) Clone() Explanation {
	e.Root = e.Root.Clone()
	return e
}
