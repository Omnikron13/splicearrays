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

// A TreeSlab is a slab-allocated array of nodes.
// It is used to store a tree structure which doesn't suffer from the fragmentation issues that a
// pointer-based tree would.
// In theory this should improve performance issues related to virtual memory management, cache misses, and
// garbage collection.
type TreeSlab struct {
	nodes []node
}

// newTreeSlab creates a new TreeSlab with an initial capacity of INITIAL_SLAB_CAPACITY.
func NewTreeSlab() TreeSlab {
	return TreeSlab{nodes: make([]node, 0, INITIAL_SLAB_CAPACITY)}
}

// ItemCount returns the total number of bytes contained in the (sub)tree rooted at the given node index.
// It also returns how many of those bytes are in the left subtree.
func (ts *TreeSlab) ItemCount(i uint32) (uint32, uint32) {
	if ts.nodes[i].leaf {
		return ts.nodes[i].y, ts.nodes[i].y
	}
	left, _ := ts.ItemCount(ts.nodes[i].x)
	right, _ := ts.ItemCount(ts.nodes[i].y)
	return left + right, left
}

// addNode adds a node to the TreeSlab.
// It returns the index of the added node.
func (ts *TreeSlab) addNode(leaf bool, x, y uint32) uint32 {
	ts.nodes = append(ts.nodes, node{leaf, x, y})
	return uint32(len(ts.nodes) - 1)
}

// addBranch adds a branch node to the TreeSlab, as a convenience method.
// It returns the index of the added node.
func (ts *TreeSlab) addBranch(left, right uint32) uint32 {
	return ts.addNode(false, left, right)
}

// AddLeaf adds a leaf node to the TreeSlab, as a convenience method.
// It returns the index of the added node.
func (ts *TreeSlab) AddLeaf(index, length uint32) uint32 {
	return ts.addNode(true, index, length)
}

// insertIntoLeaf inserts a new leaf node into the TreeSlab at given leaf node index, returning a new branch node index.
// If the insert_index is 0 the branch will have the new leaf on the left and old leaf on the right.
// If the insert_index is the length of the leaf node the branch will have the old leaf on the left and the new leaf on the right.
// If the insert_index is the length of the leaf node, a new branch node is returned with the old leaf on the left and the new leaf on the right.
// Otherwise the branch will take the form branch{split left, branch{new, split right}}.
// The index of the new branch node can be used to replace the leaf node.
func (ts *TreeSlab) insertIntoLeaf(leaf_index, insert_index, leaf uint32) uint32 {
	l, r := ts.nodes[leaf_index].remove(insert_index, 0)
	if l == nil {
		return ts.addBranch(leaf, leaf_index)
	}
	if r == nil {
		return ts.addBranch(leaf_index, leaf)
	}
	return ts.addBranch(ts.AddLeaf(l.x, l.y), ts.addBranch(leaf, ts.AddLeaf(r.x, r.y)))
}

// insertIntoBranch inserts a new leaf node into the TreeSlab at the given branch node index, returning a new branch node index.
// The new leaf node is inserted into the left or right subtree of the branch node, depending on the insert_index.
// The index of the new branch node can be used to replace the old branch node.
func (ts *TreeSlab) insertIntoBranch(branch_index, insert_index, leaf uint32) uint32 {
	_, left := ts.ItemCount(branch_index)
	println(left, insert_index)
	if insert_index < left {
		return ts.addBranch(
			ts.InsertIntoNode(ts.nodes[branch_index].x, insert_index, leaf),
			ts.nodes[branch_index].y,
		)
	}
	return ts.addBranch(
		ts.nodes[branch_index].x,
		ts.InsertIntoNode(ts.nodes[branch_index].y, insert_index-left, leaf),
	)
}

// InsertIntoNode inserts a new leaf node into the TreeSlab at the given node index, returning a new branch node index.
func (ts *TreeSlab) InsertIntoNode(node_index, insert_index, leaf uint32) uint32 {
	// TODO: technically this should actually be able to insert entire sub-trees, not just leaves
	if ts.nodes[node_index].leaf {
		return ts.insertIntoLeaf(node_index, insert_index, leaf)
	}
	return ts.insertIntoBranch(node_index, insert_index, leaf)
}

// removeFromLeaf remove a range of indices from a leaf node in the TreeSlab, returning a new leaf or branch node index,
// or nil if the leaf node is entirely removed.
func (ts *TreeSlab) removeFromLeaf(index, start, length uint32) *uint32 {
	// TODO: perhaps just have some kind of status/error return instead of mucking about with pointers?
	l, r := ts.nodes[index].remove(start, length)
	if l == nil {
		return nil
	}
	if r == nil {
		i := ts.AddLeaf(l.x, l.y)
		return &i
	}
	i := ts.addBranch(ts.AddLeaf(l.x, l.y), ts.AddLeaf(r.x, r.y))
	return &i
}

// removeFromBranch remove a range of indices from a branch node in the TreeSlab, returning a new leaf or branch node
// index, or nil if the branch node is entirely removed.
func (ts *TreeSlab) removeFromBranch(index, start, length uint32) *uint32 {
	// TODO: handle invalid inputs
	if length == 0 {
		return &index
	}

	count, left := ts.ItemCount(index)

	if start == 0 {
		// Entire range of branch removed
		if length == count {
			return nil
		}
		// Entire left side removed
		if length >= left {
			return ts.RemoveFromNode(ts.nodes[index].y, 0, length-left)
		}
	}

	// Entire right side removed
	if start <= left && start+length == count {
		return ts.RemoveFromNode(ts.nodes[index].x, start, length-(count-left))
	}

	// Contained entirely within left side
	if start+length <= left {
		i := ts.RemoveFromNode(ts.nodes[index].x, start, length)
		bi := ts.addBranch(*i, ts.nodes[index].y)
		return &bi
	}

	// Contained entirely within right side
	if start >= left {
		i := ts.RemoveFromNode(ts.nodes[index].y, start-left, length)
		bi := ts.addBranch(ts.nodes[index].x, *i)
		return &bi
	}

	li := ts.RemoveFromNode(ts.nodes[index].x, start, left-start)
	ri := ts.RemoveFromNode(ts.nodes[index].y, 0, length-(left-start))
	bi := ts.addBranch(*li, *ri)
	return &bi
}

// RemoveFromNode remove a range of indices from a node in the TreeSlab, returning a new leaf or branch node index,
// or nil if the leaf node is entirely removed.
func (ts *TreeSlab) RemoveFromNode(index, start, length uint32) *uint32 {
	if ts.nodes[index].leaf {
		return ts.removeFromLeaf(index, start, length)
	}
	return ts.removeFromBranch(index, start, length)
}

// WalkTree is a recursive function that walks the tree starting at a given index.
// It calls the given function on each node in the tree.
func (ts *TreeSlab) WalkTree(index uint32, f func(node)) {
	f(ts.nodes[index])
	if !ts.nodes[index].leaf {
		ts.WalkTree(ts.nodes[index].x, f)
		ts.WalkTree(ts.nodes[index].y, f)
	}
}

// GetLeaves returns a slice of all the leaf node indexes in the TreeSlab (sub)tree starting at a given index.
func (ts *TreeSlab) GetLeaves(index uint32) []node {
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
