package bitfield

// Bitfield is a bitfield
type Bitfield []byte

func (b Bitfield) HasPiece(index int) bool {
	byIndex := index / 8
	offset := index % 8
	return b[byIndex]>>(7-offset)&1 != 0
}

func (b Bitfield) Set(index int) {
	b[index/8] |= 1 << (7 - index%8)
}
