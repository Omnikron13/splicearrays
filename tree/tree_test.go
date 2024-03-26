package tree

import (
	"testing"
)

func TestNewTreeSlab(t *testing.T) {
	ts := NewTreeSlab()
	if len(ts.nodes) != 0 {
		t.Fail()
	}
	if cap(ts.nodes) != INITIAL_SLAB_CAPACITY {
		t.Fail()
	}
}

func TestLen(t *testing.T) {
	ts := NewTreeSlab()
	ts.AddLeaf(0, 10)
	t.Run("leaf", func(t *testing.T) {
		x := ts.Len(0)
		if x != 10 {
			t.Error("got:", x, "expected:", 10)
		}
	})

	t.Run("branch", func(t *testing.T) {
		bi := ts.addBranch(0, ts.AddLeaf(10, 20))
		x := ts.Len(bi)
		if x != 30 {
			t.Error("got:", x, "expected:", 30)
		}
	})
}

func TestByteCount(t *testing.T) {
	t.Run("Single leaf", func(t *testing.T) {
		ts := NewTreeSlab()
		li := ts.AddLeaf(0, 10)
		x, l := ts.ItemCount(li)
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
		ts := NewTreeSlab()
		i := ts.addBranch(
			ts.AddLeaf(0, 10),
			ts.AddLeaf(10, 20),
		)
		x, l := ts.ItemCount(i)
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
	ts := NewTreeSlab()
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
	ts := NewTreeSlab()
	li := ts.AddLeaf(0, 0)
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
	ts := NewTreeSlab()
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
	ts := NewTreeSlab()
	li1 := ts.AddLeaf(0, 0)
	leaves := ts.GetLeaves(li1)
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
	li2 := ts.AddLeaf(1, 1)
	bi1 := ts.addBranch(li1, li2)
	leaves = ts.GetLeaves(bi1)
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

func TestInsert(t *testing.T) {
	insert := func(setup func() (TreeSlab, uint32)) func(*testing.T, string, uint32, [][2]uint32) {
		return func(t *testing.T, name string, index uint32, expected [][2]uint32) {
			ts, ri := setup()
			root := ts.insert(ri, index, ts.AddLeaf(10, 10))
			i := 0
			for n := range ts.LeafIter(root) {
				if n.x != expected[i][0] || n.y != expected[i][1] {
					t.Errorf("leaf %d expected %d:%d, got %d:%d", i, expected[i][0], expected[i][1], n.x, n.y)
				}
				i++
			}
		}
	}
	insertLeaf := insert(func() (TreeSlab, uint32) {
		ts := NewTreeSlab()
		rootIndex := ts.AddLeaf(0, 10)
		return ts, rootIndex
	})
	insertBranch := insert(func() (TreeSlab, uint32) {
		ts := NewTreeSlab()
		rootIndex := ts.addBranch(ts.AddLeaf(0, 5), ts.AddLeaf(5, 5))
		return ts, rootIndex
	})

	t.Run("leaf", func(t *testing.T) {
		tests :=
			[]struct {
				name     string
				index    uint32
				expected [][2]uint32
			}{
				{"start", 0, [][2]uint32{{10, 10}, {0, 10}}},
				{"end", 10, [][2]uint32{{0, 10}, {10, 10}}},
				{"mid", 5, [][2]uint32{{0, 5}, {10, 10}, {5, 5}}},
			}

		for _, test := range tests {
			t.Run(test.name, func(t *testing.T) {
				insertLeaf(t, test.name, test.index, test.expected)
			})
		}
	})

	t.Run("branch", func(t *testing.T) {
		tests :=
			[]struct {
				name     string
				index    uint32
				expected [][2]uint32
			}{
				{"start", 0, [][2]uint32{{10, 10}, {0, 5}, {5, 5}}},
				{"end", 10, [][2]uint32{{0, 5}, {5, 5}, {10, 10}}},
				{"mid", 5, [][2]uint32{{0, 5}, {10, 10}, {5, 5}}},
				{"mid left", 2, [][2]uint32{{0, 2}, {10, 10}, {2, 3}, {5, 5}}},
				{"mid right", 7, [][2]uint32{{0, 5}, {5, 2}, {10, 10}, {7, 3}}},
			}

		for _, test := range tests {
			t.Run(test.name, func(t *testing.T) {
				insertBranch(t, test.name, test.index, test.expected)
			})
		}
	})
}

func TestRemoveFromLeaf(t *testing.T) {
	t.Run("all", func(t *testing.T) {
		ts := NewTreeSlab()
		ts.AddLeaf(0, 10)
		n := ts.removeFromLeaf(0, 0, 10)
		if n != nil {
			t.Fail()
		}
	})

	t.Run("start", func(t *testing.T) {
		ts := NewTreeSlab()
		ts.AddLeaf(0, 10)
		n := ts.removeFromLeaf(0, 0, 5)
		if ts.nodes[*n].String() != "leaf {index: 5 length: 5}" {
			t.Errorf("Expected leaf {index: 5 length: 5}, got %s", ts.nodes[*n].String())
		}
	})

	t.Run("end", func(t *testing.T) {
		ts := NewTreeSlab()
		ts.AddLeaf(0, 10)
		n := ts.removeFromLeaf(0, 5, 5)
		if ts.nodes[*n].String() != "leaf {index: 0 length: 5}" {
			t.Errorf("Expected leaf {index: 0 length: 5}, got %s", ts.nodes[*n].String())
		}
	})

	t.Run("middle", func(t *testing.T) {
		ts := NewTreeSlab()
		ts.AddLeaf(0, 10)
		n := ts.removeFromLeaf(0, 3, 4)
		if ts.nodes[*n].leaf {
			t.Error("Expected branch node, got leaf node")
		}
		li := ts.nodes[*n].x
		if ts.nodes[li].String() != "leaf {index: 0 length: 3}" {
			t.Errorf("Expected leaf {index: 0 length: 3}, got %s", ts.nodes[li].String())
		}
		ri := ts.nodes[*n].y
		if ts.nodes[ri].String() != "leaf {index: 7 length: 3}" {
			t.Errorf("Expected leaf {index: 7 length: 3}, got %s", ts.nodes[ri].String())
		}
	})
}

func TestRemoveFromBranch(t *testing.T) {
	ts := NewTreeSlab()
	i := ts.addBranch(
		ts.AddLeaf(0, 10),
		ts.AddLeaf(10, 10),
	)

	t.Run("all", func(t *testing.T) {
		n := ts.removeFromBranch(i, 0, 20)
		if n != nil {
			t.Error("Expected nil, got", n)
		}
	})

	t.Run("all left", func(t *testing.T) {
		n := ts.removeFromBranch(i, 0, 10)
		expected := "leaf {index: 10 length: 10}"
		actual := ts.nodes[*n].String()
		if actual != expected {
			t.Errorf("Expected %s, got %s", expected, actual)
		}
	})

	t.Run("some left", func(t *testing.T) {
		n := ts.removeFromBranch(i, 0, 5)
		if n == nil {
			t.Error("Expected non-nil, got nil")
		}
		expected := "leaf {index: 5 length: 5}"
		actual := ts.nodes[ts.nodes[*n].x].String()
		if actual != expected {
			t.Errorf("Expected left %s, got %s", expected, actual)
		}
		expected = "leaf {index: 10 length: 10}"
		actual = ts.nodes[ts.nodes[*n].y].String()
		if actual != expected {
			t.Errorf("Expected right %s, got %s", expected, actual)
		}
	})

	t.Run("middle left", func(t *testing.T) {
		n := ts.removeFromBranch(i, 3, 4)
		leaves := ts.GetLeaves(*n)
		if leaves[0].String() != "leaf {index: 0 length: 3}" ||
			leaves[1].String() != "leaf {index: 7 length: 3}" ||
			leaves[2].String() != "leaf {index: 10 length: 10}" {
			t.Fail()
		}
	})

	t.Run("first three quarters", func(t *testing.T) {
		n := ts.removeFromBranch(i, 0, 15)
		expected := "leaf {index: 15 length: 5}"
		actual := ts.nodes[*n].String()
		if ts.nodes[*n].String() != expected {
			t.Errorf("Expected %s, got %s", expected, actual)
		}
	})

	t.Run("all right", func(t *testing.T) {
		n := ts.removeFromBranch(i, 10, 10)
		expected := "leaf {index: 0 length: 10}"
		actual := ts.nodes[*n].String()
		if actual != expected {
			t.Errorf("Expected %s, got %s", expected, actual)
		}
	})

	t.Run("some right", func(t *testing.T) {
		n := ts.removeFromBranch(i, 15, 5)
		if n == nil {
			t.Error("Expected non-nil, got nil")
		}
		expected := "leaf {index: 0 length: 10}"
		actual := ts.nodes[ts.nodes[*n].x].String()
		if actual != expected {
			t.Errorf("Expected left %s, got %s", expected, actual)
		}
		expected = "leaf {index: 10 length: 5}"
		actual = ts.nodes[ts.nodes[*n].y].String()
		if actual != expected {
			t.Errorf("Expected right %s, got %s", expected, actual)
		}
	})

	t.Run("middle right", func(t *testing.T) {
		n := ts.removeFromBranch(i, 13, 4)
		leaves := ts.GetLeaves(*n)
		if leaves[0].String() != "leaf {index: 0 length: 10}" ||
			leaves[1].String() != "leaf {index: 10 length: 3}" ||
			leaves[2].String() != "leaf {index: 17 length: 3}" {
			t.Fail()
		}
	})

	t.Run("last three quarters", func(t *testing.T) {
		n := ts.removeFromBranch(i, 5, 15)
		expected := "leaf {index: 0 length: 5}"
		actual := ts.nodes[*n].String()
		if ts.nodes[*n].String() != expected {
			t.Errorf("Expected %s, got %s", expected, actual)
		}
	})

	t.Run("middle half", func(t *testing.T) {
		n := ts.removeFromBranch(i, 5, 10)
		expected := "leaf {index: 0 length: 5}"
		actual := ts.nodes[ts.nodes[*n].x].String()
		if actual != expected {
			t.Errorf("Expected left %s, got %s", expected, actual)
		}
		expected = "leaf {index: 15 length: 5}"
		actual = ts.nodes[ts.nodes[*n].y].String()
		if actual != expected {
			t.Errorf("Expected right %s, got %s", expected, actual)
		}
	})
}

func TestWalkTree(t *testing.T) {
	ts := NewTreeSlab()
	// TODO: tidy this setup, perhaps move to a helper for reuse
	nodes := make([]uint32, 0)
	for i := 0; i < 16; i++ {
		nodes = append(nodes, ts.AddLeaf(uint32(i), uint32(i)))
	}
	for i := 0; i < 16/2; i++ {
		nodes[i] = ts.addBranch(nodes[i*2], nodes[i*2+1])
	}
	for i := 0; i < 16/4; i++ {
		nodes[i] = ts.addBranch(nodes[i*2], nodes[i*2+1])
	}
	for i := 0; i < 16/8; i++ {
		nodes[i] = ts.addBranch(nodes[i*2], nodes[i*2+1])
	}
	idx := ts.addBranch(nodes[0], nodes[1])
	i := uint32(0)
	ts.WalkTree(idx, func(n node) {
		if n.leaf {
			if n.x != i {
				t.Error("Expected index", i, "got", n.x)
			}
			if n.y != i {
				t.Error("Expected length", i, "got", n.y)
			}
			i++
		}
	})
}

func TestLeafIter(t *testing.T) {
	ts := NewTreeSlab()
	i := ts.addBranch(
		ts.AddLeaf(0, 10),
		ts.addBranch(
			ts.addBranch(
				ts.AddLeaf(10, 10),
				ts.addBranch(
					ts.AddLeaf(20, 10),
					ts.AddLeaf(30, 10),
				),
			),
			ts.AddLeaf(40, 10),
		),
	)

	x := uint32(0)
	for n := range ts.LeafIter(i) {
		l, _ := n.Length()
		x += l
	}
	if x != 50 {
		t.Fail()
	}
}
