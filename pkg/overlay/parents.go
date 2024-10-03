package overlay

import (
	"github.com/goccy/go-yaml/ast"
)

type parentIndex struct {
	root ast.Node
}

// newParentIndex returns a new parentIndex, populated for the given root node.
func newParentIndex(root ast.Node) parentIndex {
	return parentIndex{
		root: root,
	}
}

func (index parentIndex) getParent(child ast.Node) ast.Node {
	return ast.Parent(index.root, child)
}
