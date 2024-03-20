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
	// TODO: just delegate to remove?
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

// remove removes a range of indices from a leaf node.
// Returns either two leaves if both sides of the range are not empty, one leaf & nil if one side is empty,
// or nil, nil if the entire range of the leaf is removed.
func (n *node) remove(start, length uint32) (*node, *node) {
	// TODO: deal with branch nodes
	// TODO: handle invalid inputs (start out of bounds, start+length out of bounds, split at 0 or n.y)
	if length == n.y {
		return nil, nil
	}
	if start == 0 {
		return &node{leaf: true, x: n.x + length, y: n.y - length}, nil
	}
	if start+length == n.y {
		return &node{leaf: true, x: n.x, y: n.y - length}, nil
	}
	return &node{leaf: true, x: n.x, y: start}, &node{leaf: true, x: n.x + start + length, y: n.y - (start + length)}
}
