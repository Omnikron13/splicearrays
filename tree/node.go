package tree

import "fmt"

// A node can either be a leaf or a branch.
// If it is a branch it contains an index to the left and right child nodes.
// If it is a leaf, it contains an index into the data slab, and the length of the sub-sequence.
type node struct {
	leaf bool
	x, y uint32
}

// String() returns a string representation of the node.
func (n *node) String() string {
	if n.leaf {
		return fmt.Sprintf("leaf {index: %d length: %d}", n.x, n.y)
	} else {
		return fmt.Sprintf("branch {left: %d right: %d}", n.x, n.y)
	}
}

// split splits a leaf node into two leaf nodes at the given index.
// If the index is 0, nil & the original node is returned.
// If the index is the length of the node, the original node & nil is returned.
// Otherwise, two new nodes are created and returned.
func (n *node) split(index uint32) (*node, *node) {
	// TODO: deal with branch nodes
	// TODO: error handling (e.g. if index > n.y)
	if index == 0 {
		return nil, n
	} else if index == n.y {
		return n, nil
	} else {
		return &node{leaf: true, x: n.x, y: index}, &node{leaf: true, x: n.x + index, y: n.y - index}
	}
}
