package bits

func NewArray(size int64) *Array {
	var b Array = make([]byte, size/8)
	return &b
}

type Array []byte

func (b *Array) Add(n int64) {
	index := n / 8
	pos := n % 8
	(*b)[index] |= 1 << pos
}

func (b *Array) Contains(n int64) bool {
	index := n / 8
	pos := n % 8
	return ((*b)[index] & (1 << pos)) != 0
}
