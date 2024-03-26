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
	nodes Slab[node]
}

// newTreeSlab creates a new TreeSlab with an initial capacity of INITIAL_SLAB_CAPACITY.
func NewTreeSlab() TreeSlab {
	ms := make(MinimalSlab[node], 0, INITIAL_SLAB_CAPACITY)
	return TreeSlab{nodes: &ms}
}

// Len returns the total number of items contained in the (sub)tree rooted at the given node index.
func (ts *TreeSlab) Len(index uint32) (length uint32) {
	for n := range ts.LeafIter(index) {
		length += n.y
	}
	return
}

// addNode adds a node to the TreeSlab.
// It returns the index of the added node.
func (ts *TreeSlab) addNode(leaf bool, x, y uint32) uint32 {
	i, _ := ts.nodes.Add(node{leaf, x, y})
	return i
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

// insert inserts a new node into the TreeSlab at the given branch node index, returning a new branch node index.
// The new node is inserted into the left or right subtree of the branch node, depending on the insert_index, or
// into the leaf node at the specified insert_index.
// The index of the new branch node can be used to replace the old branch node.
func (ts *TreeSlab) insert(root_index, insert_index, new_node_index uint32) uint32 {
	// short circuit out appending to index 0
	if insert_index == 0 {
		return ts.addBranch(new_node_index, root_index)
	}

	// short circuit out appending to index end
	if insert_index == ts.Len(root_index) {
		return ts.addBranch(root_index, new_node_index)
	}

	// short circuit out inserting into a leaf
	if ts.nodes.Get(root_index).leaf {
		l, r := ts.nodes.Get(root_index).remove(insert_index, 0)
		return ts.addBranch(ts.AddLeaf(l.x, l.y), ts.addBranch(new_node_index, ts.AddLeaf(r.x, r.y)))
	}

	// short circuit out appending to index in the middle of the two halves of a branch
	bn := ts.nodes.Get(root_index)
	l_len := ts.Len(bn.x)
	if insert_index == l_len {
		return ts.addBranch(bn.x, ts.addBranch(new_node_index, bn.y))
	}

	// we can now assume insert_index is in the left or right half of a branch.
	// In the left node would be slightly simpler, but we can adjust the insert_index for the right node.
	if insert_index < l_len {
		return ts.addBranch(ts.insert(bn.x, insert_index, new_node_index), bn.y)
	}
	return ts.addBranch(bn.x, ts.insert(bn.y, insert_index-l_len, new_node_index))
}

// Remove removes a range of indices from the specified (sub)tree, returning a new leaf or branch node index, or nil
// if the node is entirely removed.
func (ts *TreeSlab) Remove(index, start, length uint32) *uint32 {
	// TODO: handle invalid inputs
	// short circuit out if the entire node is removed
	if start == 0 && length == ts.Len(index) {
		return nil
	}

	// handle leaf nodes
	if ts.nodes.Get(index).leaf {
		l, r := ts.nodes.Get(index).remove(start, length)
		if r == nil {
			i := ts.AddLeaf(l.x, l.y)
			return &i
		}
		i := ts.addBranch(ts.AddLeaf(l.x, l.y), ts.AddLeaf(r.x, r.y))
		return &i
	}

	// removing from the right side of the branch only
	l_len := ts.Len(ts.nodes.Get(index).x)
	if start >= l_len {
		r := ts.Remove(ts.nodes.Get(index).y, start-l_len, length)
		if r == nil {
			return &ts.nodes.Get(index).x
		}
		bi := ts.addBranch(ts.nodes.Get(index).x, *r)
		return &bi
	}

	// removing from the left side of the branch only
	if start+length <= l_len {
		l := ts.Remove(ts.nodes.Get(index).x, start, length)
		if l == nil {
			return &ts.nodes.Get(index).y
		}
		bi := ts.addBranch(*l, ts.nodes.Get(index).y)
		return &bi
	}

	// removing from both sides of the branch
	li := ts.Remove(ts.nodes.Get(index).x, start, l_len-start)
	ri := ts.Remove(ts.nodes.Get(index).y, 0, length-(l_len-start))
	bi := ts.addBranch(*li, *ri)
	return &bi
}

// WalkTree is a recursive function that walks the tree starting at a given index.
// It calls the given function on each node in the tree.
func (ts *TreeSlab) WalkTree(index uint32, f func(*node)) {
	f(ts.nodes.Get(index))
	if !ts.nodes.Get(index).leaf {
		ts.WalkTree(ts.nodes.Get(index).x, f)
		ts.WalkTree(ts.nodes.Get(index).y, f)
	}
}

// TODO: remove this? rename it at least, if it really serves any purpose...
// GetLeaves returns a slice of all the leaf node indexes in the TreeSlab (sub)tree starting at a given index.
func (ts *TreeSlab) GetLeaves(index uint32) (leaves []node) {
	for n := range ts.LeafIter(index) {
		leaves = append(leaves, *n)
	}
	return
}

// LeafIter returns a channel that iterates over the leaf nodes in the (sub)tree starting at a given node index.
func (ts *TreeSlab) LeafIter(index uint32) chan *node {
	c := make(chan *node, 4)
	go func() {
		ts.WalkTree(index, func(n *node) {
			if n.leaf {
				c <- n
			}
		})
		close(c)
	}()
	return c
}

// IndexIter returns a channel that iterates over the indices represented by the leaf nodes in the (sub)tree starting
// at a given node index.
func (ts *TreeSlab) IndexIter(index uint32) chan uint32 {
	c := make(chan uint32, 64)
	go func() {
		for n := range ts.LeafIter(index) {
			for i := n.x; i < n.x+n.y; i++ {
				c <- i
			}
		}
		close(c)
	}()
	return c
}
