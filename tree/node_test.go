package tree

import "testing"

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

func TestRemove(t *testing.T) {
	t.Run("all", func(t *testing.T) {
		n := node{leaf: true, x: 0, y: 10}
		l, r := n.remove(0, 10)
		if l != nil || r != nil {
			t.Fail()
		}
	})

	t.Run("start", func(t *testing.T) {
		n := node{leaf: true, x: 0, y: 10}
		l, r := n.remove(0, 5)
		if r != nil {
			t.Fail()
		}
		if l.String() != "leaf {index: 5 length: 5}" {
			t.Errorf("Expected leaf {index: 5 length: 5}, got %s", l.String())
		}
	})

	t.Run("end", func(t *testing.T) {
		n := node{leaf: true, x: 0, y: 10}
		l, r := n.remove(5, 5)
		if r != nil {
			t.Fail()
		}
		if l.String() != "leaf {index: 0 length: 5}" {
			t.Errorf("Expected leaf {index: 0 length: 5}, got %s", l.String())
		}
	})

	t.Run("middle", func(t *testing.T) {
		n := node{leaf: true, x: 0, y: 10}
		l, r := n.remove(3, 4)
		if l.String() != "leaf {index: 0 length: 3}" {
			t.Errorf("Expected leaf {index: 0 length: 5}, got %s", l.String())
		}
		if r.String() != "leaf {index: 7 length: 3}" {
			t.Errorf("Expected leaf {index: 0 length: 5}, got %s", r.String())
		}
	})

	t.Run("split", func(t *testing.T) {
		n := node{leaf: true, x: 0, y: 10}
		l, r := n.remove(5, 0)
		if l.String() != "leaf {index: 0 length: 5}" {
			t.Errorf("Expected leaf {index: 0 length: 5}, got %s", l.String())
		}
		if r.String() != "leaf {index: 5 length: 5}" {
			t.Errorf("Expected leaf {index: 0 length: 5}, got %s", r.String())
		}
	})
}
