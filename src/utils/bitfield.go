package utils

type Bitfield struct {
	data  []byte
	nbits uint64
}

func NewBitfield(nbits uint64) Bitfield {
	nbytes := nbits / 8
	if nbits%8 != 0 {
		nbytes += 1
	}

	bf := Bitfield{
		data:  make([]byte, nbytes, nbytes),
		nbits: nbits,
	}

	return bf
}

func FromBytes(bytes []byte, nbits uint64) Bitfield {
	return Bitfield{
		data:  bytes,
		nbits: nbits,
	}
}

func (bf Bitfield) Fill() {
	for i, _ := range bf.data {
		bf.data[i] = 255
	}
}

func (bf Bitfield) Get(idx uint64) bool {
	maj := idx / 8
	min := idx % 8
	mask := uint8(128) >> min

	elem := bf.data[maj]
	return (elem & mask) > 0
}

func (bf Bitfield) Set(idx uint64, val bool) {
	maj := idx / 8
	min := idx % 8

	if val {
		mask := uint8(128) >> min
		bf.data[maj] |= mask
	} else {
		mask := ^(uint8(128) >> min)
		bf.data[maj] &= mask
	}
}

func (bf Bitfield) Data() []byte {
	return bf.data
}
