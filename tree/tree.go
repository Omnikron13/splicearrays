package tree

import (
	"unsafe"
)

// NODE_BYTE_SIZE is the size of a node in bytes.
const NODE_BYTE_SIZE = unsafe.Sizeof(node{})

// SLAB_CHUNK_SIZE is the number of bytes new tree_slabs are initialized to hold by default, as well as ideally
// the number of new bytes to allocate at a time when the slab is full.
const SLAB_CHUNK_SIZE = 4096

// INITIAL_SLAB_CAPACITY is the number of nodes new tree_slabs are initialized to hold by default.
const INITIAL_SLAB_CAPACITY = SLAB_CHUNK_SIZE / int(NODE_BYTE_SIZE)

// A treeSlab is a slab-allocated array of nodes.
// It is used to store a tree structure which doesn't suffer from the fragmentation issues that a
// pointer-based tree would.
// In theory this should improve performance issues related to virtual memory management, cache misses, and
// garbage collection.
type treeSlab struct {
	nodes []node
	root  uint32
}

// newTreeSlab creates a new treeSlab with an initial capacity of INITIAL_SLAB_CAPACITY.
func newTreeSlab() treeSlab {
	return treeSlab{nodes: make([]node, 0, INITIAL_SLAB_CAPACITY)}
}

// byteCount returns the total number of bytes contained in the (sub)tree rooted at the given node index.
// It also returns how many of those bytes are in the left subtree.
func (ts *treeSlab) byteCount(i uint32) (uint32, uint32) {
	if ts.nodes[i].leaf {
		return ts.nodes[i].y, ts.nodes[i].y
	}
	left, _ := ts.byteCount(ts.nodes[i].x)
	right, _ := ts.byteCount(ts.nodes[i].y)
	return left + right, left
}

// addNode adds a node to the treeSlab.
// It returns the index of the added node.
func (ts *treeSlab) addNode(leaf bool, x, y uint32) uint32 {
	ts.nodes = append(ts.nodes, node{leaf, x, y})
	return uint32(len(ts.nodes) - 1)
}

// addBranch adds a branch node to the treeSlab, as a convenience method.
// It returns the index of the added node.
func (ts *treeSlab) addBranch(left, right uint32) uint32 {
	return ts.addNode(false, left, right)
}

// addLeaf adds a leaf node to the treeSlab, as a convenience method.
// It returns the index of the added node.
func (ts *treeSlab) addLeaf(index, length uint32) uint32 {
	return ts.addNode(true, index, length)
}

// insertIntoLeaf inserts a new leaf node into the treeSlab at given leaf node index, returning a new branch node index.
// If the insert_index is 0 the branch will have the new leaf on the left and old leaf on the right.
// If the insert_index is the length of the leaf node the branch will have the old leaf on the left and the new leaf on the right.
// If the insert_index is the length of the leaf node, a new branch node is returned with the old leaf on the left and the new leaf on the right.
// Otherwise the branch will take the form branch{split left, branch{new, split right}}.
// The index of the new branch node can be used to replace the leaf node.
func (ts *treeSlab) insertIntoLeaf(leaf_index, insert_index, leaf uint32) uint32 {
	l, r := ts.nodes[leaf_index].split(insert_index)
	if l == nil {
		return ts.addBranch(leaf, leaf_index)
	}
	if r == nil {
		return ts.addBranch(leaf_index, leaf)
	}
	return ts.addBranch(ts.addLeaf(l.x, l.y), ts.addBranch(leaf, ts.addLeaf(r.x, r.y)))
}

// insertIntoBranch inserts a new leaf node into the treeSlab at the given branch node index, returning a new branch node index.
// The new leaf node is inserted into the left or right subtree of the branch node, depending on the insert_index.
// The index of the new branch node can be used to replace the old branch node.
func (ts *treeSlab) insertIntoBranch(branch_index, insert_index, leaf uint32) uint32 {
	_, left := ts.byteCount(ts.nodes[branch_index].x)
	if insert_index < left {
		return ts.addBranch(
			ts.insertIntoNode(ts.nodes[branch_index].x, insert_index, leaf),
			ts.nodes[branch_index].y,
		)
	}
	return ts.addBranch(
		ts.nodes[branch_index].x,
		ts.insertIntoNode(ts.nodes[branch_index].y, insert_index-left, leaf),
	)
}

// insertIntoNode inserts a new leaf node into the treeSlab at the given node index, returning a new branch node index.
func (ts *treeSlab) insertIntoNode(node_index, insert_index, leaf uint32) uint32 {
	if ts.nodes[node_index].leaf {
		return ts.insertIntoLeaf(node_index, insert_index, leaf)
	}
	return ts.insertIntoBranch(node_index, insert_index, leaf)
}

// getLeaves returns a slice of all the leaf node indexes in the treeSlab (sub)tree starting at a given index.
func (ts *treeSlab) getLeaves(index uint32) []node {
	// TODO: return pointers to nodes instead of copying them?
	// TODO: short circuit out for 0 & 1 total nodes
	leaves := make([]node, 0, len(ts.nodes)/2+1)

	var walkTree func(uint32)
	walkTree = func(i uint32) {
		if ts.nodes[i].leaf {
			leaves = append(leaves, ts.nodes[i])
		} else {
			walkTree(ts.nodes[i].x)
			walkTree(ts.nodes[i].y)
		}
	}
	walkTree(index)

	return leaves
}
