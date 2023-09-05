package lru

// A ByteView holds an immutable view of bytes
type ByteView struct {
	b []byte
}

// returns the view's len
func (v ByteView) Len() int {
	return len(v.b)
}

// returns the data as a string, making a copy is necessary
func (v ByteView) String() string {
	return string(v.b)
}

// returns a copy of the data as a byte slice.
func (v ByteView) ByteSlices() []byte {
	return cloneBytes(v.b)
}

func cloneBytes(b []byte) []byte {
	c := make([]byte, len(b))
	copy(c, b)
	return c
}
