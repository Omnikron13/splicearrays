package tree

import (
	"testing"
)

func TestNodeSplit(t *testing.T) {
	bn := node{x: 0, y: 10}
	left, right := bn.split(2)
	if left == nil || right == nil {
		t.Fail()
	}
	if *left != (node{leaf: true, x: 0, y: 2}) {
		t.Fail()
	}
	if *right != (node{leaf: true, x: 2, y: 8}) {
		t.Fail()
	}
	left, right = bn.split(0)
	if left != nil || *right != bn {
		t.Fail()
	}
	left, right = bn.split(10)
	if *left != bn || right != nil {
		t.Fail()
	}
}

func TestNewTreeSlab(t *testing.T) {
	ts := newTreeSlab()
	if len(ts.nodes) != 0 {
		t.Fail()
	}
	if cap(ts.nodes) != INITIAL_SLAB_CAPACITY {
		t.Fail()
	}
	if ts.root != 0 {
		t.Fail()
	}
}

func TestByteCount(t *testing.T) {
	t.Run("Single leaf", func(t *testing.T) {
		ts := newTreeSlab()
		li := ts.addLeaf(0, 10)
		x, l := ts.byteCount(li)
		t.Run("total", func(t *testing.T) {
			if x != 10 {
				t.Error("got:", x, "expected:", 10)
			}
		})

		t.Run("left", func(t *testing.T) {
			if l != 10 {
				t.Error("got:", l, "expected:", 10)
			}
		})
	})

	t.Run("Branch with two leaves", func(t *testing.T) {
		ts := newTreeSlab()
		i := ts.addBranch(
			ts.addLeaf(0, 10),
			ts.addLeaf(10, 20),
		)
		x, l := ts.byteCount(i)
		t.Run("total", func(t *testing.T) {
			if x != 30 {
				t.Error("got:", x, "expected:", 30)
			}
		})
		t.Run("left", func(t *testing.T) {
			if l != 10 {
				t.Error("got:", l, "expected:", 10)
			}
		})
	})
}

func TestAddNode(t *testing.T) {
	ts := newTreeSlab()
	bi := ts.addNode(false, 0, 0)
	if len(ts.nodes) != 1 {
		t.Fail()
	}
	if ts.nodes[0].leaf != false {
		t.Fail()
	}
	if ts.nodes[0].x != 0 {
		t.Fail()
	}
	if ts.nodes[0].y != 0 {
		t.Fail()
	}
	if bi != uint32(len(ts.nodes))-1 {
		t.Fail()
	}
	li := ts.addNode(true, 0, 0)
	if len(ts.nodes) != 2 {
		t.Fail()
	}
	if ts.nodes[li].leaf != true {
		t.Fail()
	}
	if ts.nodes[li].x != 0 {
		t.Fail()
	}
	if ts.nodes[li].y != 0 {
		t.Fail()
	}
	if li != uint32(len(ts.nodes))-1 {
		t.Fail()
	}
}

func TestAddLeaf(t *testing.T) {
	ts := newTreeSlab()
	li := ts.addLeaf(0, 0)
	if len(ts.nodes) != 1 {
		t.Fail()
	}
	if ts.nodes[li].leaf != true {
		t.Fail()
	}
	if ts.nodes[li].x != 0 {
		t.Fail()
	}
	if ts.nodes[li].y != 0 {
		t.Fail()
	}
	if li != uint32(len(ts.nodes))-1 {
		t.Fail()
	}
}

func TestAddBranch(t *testing.T) {
	ts := newTreeSlab()
	bi := ts.addBranch(0, 0)
	if len(ts.nodes) != 1 {
		t.Fail()
	}
	if ts.nodes[bi].leaf != false {
		t.Fail()
	}
	if ts.nodes[bi].x != 0 {
		t.Fail()
	}
	if ts.nodes[bi].y != 0 {
		t.Fail()
	}
	if bi != uint32(len(ts.nodes))-1 {
		t.Fail()
	}
}

func TestGetLeaves(t *testing.T) {
	ts := newTreeSlab()
	li1 := ts.addLeaf(0, 0)
	leaves := ts.getLeaves(li1)
	if len(leaves) != 1 {
		t.Fail()
	}
	if leaves[0].leaf != true {
		t.Fail()
	}
	if leaves[0].x != 0 {
		t.Fail()
	}
	if leaves[0].y != 0 {
		t.Fail()
	}
	li2 := ts.addLeaf(1, 1)
	bi1 := ts.addBranch(li1, li2)
	ts.root = bi1
	leaves = ts.getLeaves(bi1)
	if len(leaves) != 2 {
		t.Fail()
	}
	if leaves[0].leaf != true {
		t.Fail()
	}
	if leaves[0].x != 0 {
		t.Fail()
	}
	if leaves[0].y != 0 {
		t.Fail()
	}
	if leaves[1].leaf != true {
		t.Fail()
	}
	if leaves[1].x != 1 {
		t.Fail()
	}
	if leaves[1].y != 1 {
		t.Fail()
	}
}

func TestInsertIntoLeaf(t *testing.T) {
	ts := newTreeSlab()
	leaf := ts.addLeaf(0, 10)

	t.Run("Insert into leaf at 0", func(t *testing.T) {
		i := ts.insertIntoLeaf(leaf, 0, 10, 2)
		node := ts.nodes[i]
		if ts.nodes[i].String() != "branch {left: 1 right: 0}" {
			t.Fail()
		}
		if ts.nodes[node.x].String() != "leaf {index: 10 length: 2}" {
			t.Fail()
		}
		if ts.nodes[node.y].String() != "leaf {index: 0 length: 10}" {
			t.Fail()
		}
	})

	t.Run("Insert into leaf at end", func(t *testing.T) {
		i := ts.insertIntoLeaf(leaf, 10, 10, 2)
		node := ts.nodes[i]
		if ts.nodes[i].String() != "branch {left: 0 right: 3}" {
			t.Fatal("got:", ts.nodes[i].String(), "expected:", "branch {left: 0 right: 3}")
		}
		if ts.nodes[node.x].String() != "leaf {index: 0 length: 10}" {
			t.Fatal("got:", ts.nodes[node.x].String(), "expected:", "leaf {index: 0 length: 10}")
		}
		if ts.nodes[node.y].String() != "leaf {index: 10 length: 2}" {
			t.Fatal("got:", ts.nodes[node.y].String(), "expected:", "leaf {index: 10 length: 2}")
		}
	})

	t.Run("Insert into leaf at 2", func(t *testing.T) {
		i := ts.insertIntoLeaf(leaf, 2, 10, 2)
		node := ts.nodes[i]
		if ts.nodes[i].String() != "branch {left: 5 right: 8}" {
			t.Error("got:", ts.nodes[i].String(), "expected:", "branch {left: 5 right: 8}")
		}
		if ts.nodes[node.x].String() != "leaf {index: 0 length: 2}" {
			t.Error("got:", ts.nodes[node.x].String(), "expected:", "leaf {index: 0 length: 2}")
		}
		if ts.nodes[node.y].String() != "branch {left: 6 right: 7}" {
			t.Error("got:", ts.nodes[node.y].String(), "expected:", "branch {left: 6 right: 7}")
		}
		node = ts.nodes[node.y]
		if ts.nodes[node.x].String() != "leaf {index: 10 length: 2}" {
			t.Error("got:", ts.nodes[node.x].String(), "expected:", "leaf {index: 10 length: 2}")
		}
		if ts.nodes[node.y].String() != "leaf {index: 2 length: 8}" {
			t.Error("got:", ts.nodes[node.y].String(), "expected:", "leaf {index: 2 length: 8}")
		}
	})
}
