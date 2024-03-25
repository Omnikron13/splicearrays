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

// Len returns the total number of items contained in the (sub)tree rooted at the given node index.
func (ts *TreeSlab) Len(index uint32) uint32 {
	len := uint32(0)
	for n := range ts.LeafIter(index) {
		len += n.y
	}
	return len
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
	if ts.nodes[root_index].leaf {
		l, r := ts.nodes[root_index].remove(insert_index, 0)
		return ts.addBranch(ts.AddLeaf(l.x, l.y), ts.addBranch(new_node_index, ts.AddLeaf(r.x, r.y)))
	}

	// short circuit out appending to index in the middle of the two halves of a branch
	bn := ts.nodes[root_index]
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

// LeafIter returns a channel that iterates over the leaf nodes in the (sub)tree starting at a given index.
func (ts *TreeSlab) LeafIter(index uint32) chan node {
	c := make(chan node, 1)
	go func() {
		ts.WalkTree(index, func(n node) {
			if n.leaf {
				c <- n
			}
		})
		close(c)
	}()
	return c
}
