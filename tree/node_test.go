package tree

import "testing"

func TestNode(t *testing.T) {
	bn := node{x: 0, y: 0}
	if bn.leaf != false {
		t.Fail()
	}
	if bn.x != 0 {
		t.Fail()
	}
	if bn.y != 0 {
		t.Fail()
	}
	ln := node{leaf: true, x: 0, y: 0}
	if ln.leaf != true {
		t.Fail()
	}
	if ln.x != 0 {
		t.Fail()
	}
	if ln.y != 0 {
		t.Fail()
	}
}
func TestNodeToString(t *testing.T) {
	t.Run("branch", func(t *testing.T) {
		n := node{x: 1, y: 2}
		str := n.String()
		expect := "branch {left: 1 right: 2}"
		if str != expect {
			t.Errorf("Expected %s, got %s", expect, str)
		}
	})

	t.Run("leaf", func(t *testing.T) {
		n := node{leaf: true, x: 3, y: 4}
		str := n.String()
		expect := "leaf {index: 3 length: 4}"
		if str != expect {
			t.Errorf("Expected %s, got %s", expect, str)
		}
	})
}
