package ycache

type ByteView struct {
	data []byte
}

func (b ByteView) Len() int {
	return len(b.data)
}

func (b ByteView) String() string {
	return string(b.data)
}

func cloneBytes(b []byte) []byte {
	c := make([]byte, len(b))
	copy(c, b)
	return c
}

func (b ByteView) ByteSlice() []byte {
	return cloneBytes(b.data)
}
