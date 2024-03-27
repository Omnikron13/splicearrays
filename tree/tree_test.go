package tree

import (
	"math/rand"
	"testing"
)

// generates a balanced tree with 2^depth leaves with length 2^width.
// Note that if width + depth > 32, the tree will explode.
func generateBalancedTree(depth uint32, width uint32) (ts TreeSlab, root uint32) {
	// TODO: perhaps try to avoid exploding with depth over (or near) 32...
	ts = NewTreeSlab()
	depth = 1 << depth
	width = 1 << width
	nodes := make([]uint32, 0, depth)
	for i := uint32(0); i < depth; i++ {
		nodes = append(nodes, ts.AddLeaf(i*width, width))
	}
	for x := depth; x > 1; x /= 2 {
		for i := uint32(0); i < x/2; i++ {
			nodes[i] = ts.addBranch(nodes[i*2], nodes[i*2+1])
		}
	}
	root = nodes[0]
	return
}

// generateUnbalancedTree generates a tree with 2^depth leaves with length 2^width.
// The level of imbalance is random, due to the method of generating the tree from the bottom up.
// The skew parameter controls the direction of the imbalance of the last abs(skew) branches, with
// negative skew favoring the left branch and positive skew favoring the right.
func generateUnbalancedTree(depth uint32, width uint32, skew int) (ts TreeSlab, root uint32) {
	ts = NewTreeSlab()
	depth = 1 << depth
	width = 1 << width
	var skewLeft bool
	if skew < 0 {
		skew = -skew
		skewLeft = true
	}
	nodes := make([]uint32, depth)
	for i := range nodes {
		nodes[i] = ts.AddLeaf(uint32(i)*width, width)
	}
	for ; len(nodes) > 1+skew; nodes = nodes[1:] {
		rand.Shuffle(len(nodes), func(i, j int) {
			nodes[i], nodes[j] = nodes[j], nodes[i]
		})
		nodes[1] = ts.addBranch(nodes[0], nodes[1])
	}
	for ; len(nodes) > 1; nodes = nodes[1:] {
		if ts.Len(nodes[0]) < ts.Len(nodes[1]) && skewLeft {
			nodes[0], nodes[1] = nodes[1], nodes[0]
		}
		if ts.Len(nodes[0]) > ts.Len(nodes[1]) && !skewLeft {
			nodes[0], nodes[1] = nodes[1], nodes[0]
		}
		nodes[1] = ts.addBranch(nodes[0], nodes[1])
	}
	root = nodes[0]
	return
}

func TestNewTreeSlab(t *testing.T) {
	ts := NewTreeSlab()
	if ts.nodes.Len() != 0 {
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

func TestAddNode(t *testing.T) {
	ts := NewTreeSlab()
	bi := ts.addNode(false, 0, 0)
	if ts.nodes.Len() != 1 {
		t.Fail()
	}
	if ts.nodes.Get(0).leaf != false {
		t.Fail()
	}
	if ts.nodes.Get(0).x != 0 {
		t.Fail()
	}
	if ts.nodes.Get(0).y != 0 {
		t.Fail()
	}
	if bi != ts.nodes.Len()-1 {
		t.Fail()
	}
	li := ts.addNode(true, 0, 0)
	if ts.nodes.Len() != 2 {
		t.Fail()
	}
	if ts.nodes.Get(li).leaf != true {
		t.Fail()
	}
	if ts.nodes.Get(li).x != 0 {
		t.Fail()
	}
	if ts.nodes.Get(li).y != 0 {
		t.Fail()
	}
	if li != ts.nodes.Len()-1 {
		t.Fail()
	}
}

func TestAddLeaf(t *testing.T) {
	ts := NewTreeSlab()
	li := ts.AddLeaf(0, 0)
	if ts.nodes.Len() != 1 {
		t.Fail()
	}
	if ts.nodes.Get(li).leaf != true {
		t.Fail()
	}
	if ts.nodes.Get(li).x != 0 {
		t.Fail()
	}
	if ts.nodes.Get(li).y != 0 {
		t.Fail()
	}
	if li != ts.nodes.Len()-1 {
		t.Fail()
	}
}

func TestAddBranch(t *testing.T) {
	ts := NewTreeSlab()
	bi := ts.addBranch(0, 0)
	if ts.nodes.Len() != 1 {
		t.Fail()
	}
	if ts.nodes.Get(bi).leaf != false {
		t.Fail()
	}
	if ts.nodes.Get(bi).x != 0 {
		t.Fail()
	}
	if ts.nodes.Get(bi).y != 0 {
		t.Fail()
	}
	if bi != ts.nodes.Len()-1 {
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

func TestTreeRemove(t *testing.T) {
	t.Run("leaf", func(t *testing.T) {
		ts := NewTreeSlab()
		ts.AddLeaf(0, 10)

		t.Run("all", func(t *testing.T) {
			ni := ts.Remove(0, 0, 10)
			if ni != nil {
				n := ts.nodes.Get(*ni)
				t.Error("Expected nil, got", *ni, "-", n.String())
			}
		})

		t.Run("start", func(t *testing.T) {
			ni := ts.Remove(0, 0, 5)
			n := ts.nodes.Get(*ni)
			if n.String() != "leaf {index: 5 length: 5}" {
				t.Errorf("Expected leaf {index: 5 length: 5}, got %s", n.String())
			}
		})

		t.Run("end", func(t *testing.T) {
			ni := ts.Remove(0, 5, 5)
			n := ts.nodes.Get(*ni)
			if n.String() != "leaf {index: 0 length: 5}" {
				t.Errorf("Expected leaf {index: 0 length: 5}, got %s", n.String())
			}
		})

		t.Run("middle", func(t *testing.T) {
			n := ts.Remove(0, 3, 4)
			if ts.nodes.Get(*n).leaf {
				t.Error("Expected branch node, got leaf node")
			}
			li := ts.nodes.Get(*n).x
			ri := ts.nodes.Get(*n).y
			l := ts.nodes.Get(li)
			r := ts.nodes.Get(ri)
			if l.String() != "leaf {index: 0 length: 3}" {
				t.Errorf("Expected leaf {index: 0 length: 3}, got %s", l.String())
			}
			if r.String() != "leaf {index: 7 length: 3}" {
				t.Errorf("Expected leaf {index: 7 length: 3}, got %s", r.String())
			}
		})
	})

	t.Run("branch", func(t *testing.T) {
		ts := NewTreeSlab()
		root := ts.addBranch(
			ts.AddLeaf(0, 10),
			ts.AddLeaf(10, 10),
		)

		t.Run("all", func(t *testing.T) {
			ni := ts.Remove(root, 0, 20)
			if ni != nil {
				n := ts.nodes.Get(*ni)
				t.Error("Expected nil, got", *ni, "-", n.String())
			}
		})

		t.Run("all left", func(t *testing.T) {
			ni := ts.Remove(root, 0, 10)
			n := ts.nodes.Get(*ni)
			if n.String() != "leaf {index: 10 length: 10}" {
				t.Errorf("Expected leaf {index: 10 length: 10}, got %s", n.String())
			}
		})

		t.Run("all right", func(t *testing.T) {
			ni := ts.Remove(root, 10, 10)
			n := ts.nodes.Get(*ni)
			if n.String() != "leaf {index: 0 length: 10}" {
				t.Errorf("Expected leaf {index: 0 length: 10}, got %s", n.String())
			}
		})

		t.Run("some left", func(t *testing.T) {
			n := ts.Remove(root, 0, 5)
			if ts.nodes.Get(*n).leaf {
				t.Error("Expected branch node, got leaf node")
			}
			l := ts.nodes.Get(ts.nodes.Get(*n).x)
			r := ts.nodes.Get(ts.nodes.Get(*n).y)
			if l.String() != "leaf {index: 5 length: 5}" {
				t.Errorf("Expected left leaf {index: 5 length: 5}, got %s", l.String())
			}
			if r.String() != "leaf {index: 10 length: 10}" {
				t.Errorf("Expected right leaf {index: 10 length: 10}, got %s", r.String())
			}
		})

		t.Run("some right", func(t *testing.T) {
			n := ts.Remove(root, 15, 5)
			if ts.nodes.Get(*n).leaf {
				t.Error("Expected branch node, got leaf node")
			}
			l := ts.nodes.Get(ts.nodes.Get(*n).x)
			r := ts.nodes.Get(ts.nodes.Get(*n).y)
			if l.String() != "leaf {index: 0 length: 10}" {
				t.Errorf("Expected left leaf {index: 0 length: 10}, got %s", l.String())
			}
			if r.String() != "leaf {index: 10 length: 5}" {
				t.Errorf("Expected right leaf {index: 10 length: 5}, got %s", r.String())
			}
		})

		t.Run("middle left", func(t *testing.T) {
			n := ts.Remove(root, 3, 4)
			leaves := ts.GetLeaves(*n)
			if leaves[0].String() != "leaf {index: 0 length: 3}" ||
				leaves[1].String() != "leaf {index: 7 length: 3}" ||
				leaves[2].String() != "leaf {index: 10 length: 10}" {
				t.Fail()
			}
		})

		t.Run("middle right", func(t *testing.T) {
			n := ts.Remove(root, 13, 4)
			leaves := ts.GetLeaves(*n)
			if leaves[0].String() != "leaf {index: 0 length: 10}" ||
				leaves[1].String() != "leaf {index: 10 length: 3}" ||
				leaves[2].String() != "leaf {index: 17 length: 3}" {
				t.Fail()
			}
		})

		t.Run("middle", func(t *testing.T) {
			n := ts.Remove(root, 5, 10)
			leaves := ts.GetLeaves(*n)
			if leaves[0].String() != "leaf {index: 0 length: 5}" ||
				leaves[1].String() != "leaf {index: 15 length: 5}" {
				t.Fail()
			}
		})
	})
}

func TestWalkTree(t *testing.T) {
	ts, idx := generateBalancedTree(4, 0)
	i := uint32(0)
	ts.WalkTree(idx, func(n *node) {
		if n.leaf {
			if n.x != i {
				t.Error("Expected index", i, "got", n.x)
			}
			if n.y != 1 {
				t.Error("Expected length 1 got", n.y)
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

func TestIndexIter(t *testing.T) {
	ts, idx := generateBalancedTree(4, 4)
	i := uint32(0)
	for n := range ts.IndexIter(idx) {
		if n != i {
			t.Error("Expected", i, "got", n)
		}
		i++
	}
}

func BenchmarkIndexIter(b *testing.B) {
	b.Run("balanced deep", func(b *testing.B) {
		ts, idx := generateBalancedTree(12, 4)
		b.ResetTimer()
		for n := 0; n < b.N; n++ {
			for i := range ts.IndexIter(idx) {
				_ = i
			}
		}
	})

	b.Run("balanced wide", func(b *testing.B) {
		ts, idx := generateBalancedTree(4, 12)
		b.ResetTimer()
		for n := 0; n < b.N; n++ {
			for i := range ts.IndexIter(idx) {
				_ = i
			}
		}
	})

   b.Run("unbalanced random", func(b *testing.B) {
      ts, idx := generateUnbalancedTree(12, 4, 0)
      b.ResetTimer()
      for n := 0; n < b.N; n++ {
         for i := range ts.IndexIter(idx) {
            _ = i
         }
      }
   })

   b.Run("unbalanced left", func(b *testing.B) {
      ts, idx := generateUnbalancedTree(12, 4, -2)
      b.ResetTimer()
      for n := 0; n < b.N; n++ {
         for i := range ts.IndexIter(idx) {
            _ = i
         }
      }
   })

   b.Run("unbalanced right", func(b *testing.B) {
      ts, idx := generateUnbalancedTree(12, 4, 2)
      b.ResetTimer()
      for n := 0; n < b.N; n++ {
         for i := range ts.IndexIter(idx) {
            _ = i
         }
      }
   })
}
