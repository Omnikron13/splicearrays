package tree

// Slab is the interface for all the slab implementations in this package.
type Slab[T any] interface {
	// Add adds items to the slab and returns the start index and number of added items.
	Add(items ...T) (start uint32, length uint32)
	// Get returns the item at the given index.
	Get(index uint32) T
	// Len returns the total number of items in the slab.
	Len() uint32
	// SliceIter returns a channel that iterates over length items from the start index.
	SliceIter(start uint32, length uint32) chan T
}

// MinimalSlab is the bare minimum implementation of a Slab.
// It is primarily intended to be used as a placeholder, for testing, benchmarking, etc.
type MinimalSlab[T any] []T

func (s *MinimalSlab[T]) Add(items ...T) (start uint32, length uint32) {
	start = uint32(len(*s))
	length = uint32(len(items))
	*s = append(*s, items...)
	return
}

func (s MinimalSlab[T]) Get(index uint32) T {
	return s[index]
}

func (s MinimalSlab[T]) Len() uint32 {
	return uint32(len(s))
}

func (s MinimalSlab[T]) SliceIter(start uint32, end uint32) chan T {
   // TODO Error handling
   length := end - start
	c := make(chan T, 1)
	go func() {
		for i := start; i < start+length; i++ {
			c <- s.Get(i)
		}
		close(c)
	}()
	return c
}

