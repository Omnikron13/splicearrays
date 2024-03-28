package tree

import "fmt"

// A node can either be a leaf or a branch.
// If it is a branch it contains an index to the left and right child nodes.
// If it is a leaf, it contains an index into the data slab, and the length of the sub-sequence.
type node struct {
	leaf bool
	x, y uint32
}

// Left returns the left child node index of a branch node. If the node is a leaf, it returns 0, and an error.
func (n *node) Left() (uint32, error) {
	if n.leaf {
		return 0, fmt.Errorf("node is a leaf")
	}
	return n.x, nil
}

// Right returns the right child node index of a branch node. If the node is a leaf, it returns 0, and an error.
func (n *node) Right() (uint32, error) {
	if n.leaf {
		return 0, fmt.Errorf("node is a leaf")
	}
	return n.y, nil
}

// Start returns the starting index for the sequence a leaf node.
// It will return 0 and an error for branch nodes, where there isn't a meaningful response.
func (n *node) Start() (uint32, error) {
	if n.leaf {
		return n.x, nil
	}
	return 0, fmt.Errorf("node is a branch")
}

// Length returns the length of the sequence stored by a leaf node.
// It will return 0 and an error for branch nodes, where there isn't a meaningful response.
func (n *node) Length() (uint32, error) {
	if n.leaf {
		return n.y, nil
	}
	return 0, fmt.Errorf("node is a branch")
}

// String() returns a string representation of the node.
func (n *node) String() string {
	if n.leaf {
		return fmt.Sprintf("leaf {index: %d length: %d}", n.x, n.y)
	} else {
		return fmt.Sprintf("branch {left: %d right: %d}", n.x, n.y)
	}
}

// remove removes a range of indices from a leaf node.
// Returns either two leaves if both sides of the range are not empty, one leaf & nil if one side is empty,
// or nil, nil if the entire range of the leaf is removed.
func (n node) remove(start, length uint32) (*node, *node) {
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
