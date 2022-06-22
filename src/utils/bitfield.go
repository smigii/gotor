package utils

import "fmt"

type Bitfield struct {
	data  []byte
	nbits uint64
	nset  uint64
}

func NewBitfield(nbits uint64) *Bitfield {
	nbytes := nbits / 8
	if nbits%8 != 0 {
		nbytes += 1
	}

	bf := Bitfield{
		data:  make([]byte, nbytes, nbytes),
		nbits: nbits,
		nset:  0,
	}

	return &bf
}

func FromBytes(bytes []byte, nbits uint64) (*Bitfield, error) {
	min := uint64((len(bytes)-1)*8) + 1
	max := uint64(len(bytes) * 8)
	if nbits < min || nbits > max {
		return nil, fmt.Errorf("invalid number of bits specified for byte slice of len %v\nmin %v, max %v, given %v", len(bytes), min, max, nbits)
	}

	bf := Bitfield{
		data:  bytes,
		nbits: nbits,
		nset:  0,
	}
	bf.calcNset()

	return &bf, nil
}

// calcNset goes through all the bytes and counts how many bits are set.
// Overwrites Bitfield.nset accordingly.
func (bf *Bitfield) calcNset() {
	var i uint64
	const maskStart = uint8(128)
	nbytes := uint64(len(bf.data))
	rem := bf.nbits % 8
	nset := uint64(0)

	// If nbits is multiple of 8, we can just read through
	// all the bytes.
	if rem == 0 {
		nbytes++
	}

	for i = 0; i < nbytes-1; i++ {
		for j := 0; j < 8; j++ {
			mask := maskStart >> j
			if (bf.data[i] & mask) > 0 {
				nset++
			}
		}
	}

	// If nbits was not a multiple of 8, we need to read the last few bits
	// of the last byte, making sure to not overcount. This will be skipped
	// if nbits % 8 == 0
	for i = 0; i < rem; i++ {
		mask := maskStart >> i
		if (bf.data[nbytes-1] & mask) > 0 {
			nset++
		}
	}

	bf.nset = nset
}

func (bf *Bitfield) Fill() {
	for i, _ := range bf.data {
		bf.data[i] = 255
	}
	bf.nset = bf.nbits
}

func (bf *Bitfield) Get(idx uint64) bool {
	maj := idx / 8
	min := idx % 8
	mask := uint8(128) >> min

	elem := bf.data[maj]
	return (elem & mask) > 0
}

func (bf *Bitfield) Set(idx uint64, val bool) {
	maj := idx / 8
	min := idx % 8
	getMask := uint8(128) >> min
	isSet := (bf.data[maj] & getMask) > 0

	if !isSet && val {
		mask := uint8(128) >> min
		bf.data[maj] |= mask
		bf.nset++
	} else if isSet && !val {
		mask := ^(uint8(128) >> min)
		bf.data[maj] &= mask
		bf.nset--
	}
}

func (bf *Bitfield) Data() []byte {
	return bf.data
}

func (bf *Bitfield) Complete() bool {
	return bf.nset == bf.nbits
}
