package bf

import "fmt"

// Bitfield is fairly self-explanatory, aside from the data5 field. Rather than
// only holding the data of the bitfield, the first 5 bytes represent the length
// (4 bytes) and id (1 byte) of a bitfield message. This is so we are not
// wasting time prepending 5 bytes to a potentially long slice every time we need
// to send a bitfield message.
type Bitfield struct {
	data5  []byte // Holds the first 5 bytes + bitfield data
	data   []byte // Holds only the bitfield data
	nbytes int64
	nbits  int64
	nset   int64
}

func NewBitfield(nbits int64) *Bitfield {
	nbytes := nbits / 8
	if nbits%8 != 0 {
		nbytes += 1
	}

	data5 := make([]byte, 5+nbytes, 5+nbytes)

	bf := Bitfield{
		data5:  data5,
		data:   data5[5:],
		nbits:  nbits,
		nbytes: nbytes,
		nset:   0,
	}

	return &bf
}

// FromBytes returns a Bitfield from the bytes provided in bytes5. bytes5 MUST
// contain the 5 (len+id) bytes, so that we are not wasting time prepending a
// potentially long array with 5 bytes.
func FromBytes(bytes5 []byte, nbits int64) (*Bitfield, error) {
	lenbytes := int64(len(bytes5))
	if lenbytes <= 5 {
		return nil, fmt.Errorf("bytes5 must have length greater than 5 bytes")
	}

	min := ((lenbytes - 5 - 1) * 8) + 1
	max := (lenbytes - 5) * 8
	if nbits < min || nbits > max {
		return nil, fmt.Errorf("invalid number of bits specified for byte slice of len %v\nmin %v, max %v, given %v", len(bytes5), min, max, nbits)
	}

	bf := Bitfield{
		data5:  bytes5,
		data:   bytes5[5:],
		nbits:  nbits,
		nbytes: lenbytes - 5,
		nset:   0,
	}
	bf.calcNset()

	return &bf, nil
}

// calcNset goes through all the bytes and counts how many bits are set.
// Overwrites Bitfield.nset accordingly.
func (bf *Bitfield) calcNset() {
	var i int64
	const maskStart = uint8(128)
	last := int64(len(bf.data) - 1)
	rem := bf.nbits % 8
	nset := int64(0)

	// If nbits is multiple of 8, we can just read through
	// all the bytes.
	if rem == 0 {
		last++
	}

	for i = 0; i < last; i++ {
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
		if (bf.data[last] & mask) > 0 {
			nset++
		}
	}

	bf.nset = nset
}

func (bf *Bitfield) Fill() {
	for i, _ := range bf.data5 {
		bf.data5[i] = 255
	}
	bf.nset = bf.nbits
}

func (bf *Bitfield) Get(idx int64) bool {
	maj := idx / 8
	min := idx % 8
	mask := uint8(128) >> min

	elem := bf.data[maj]
	return (elem & mask) > 0
}

func (bf *Bitfield) Set(idx int64, val bool) {
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

// Data5 returns the entire bitfield, including the first 5 bytes which
// represent the length (4 bytes) and id (1 byte)
func (bf *Bitfield) Data5() []byte {
	return bf.data5
}

func (bf *Bitfield) Data() []byte {
	return bf.data
}

func (bf *Bitfield) Complete() bool {
	return bf.nset == bf.nbits
}

func (bf *Bitfield) Nbits() int64 {
	return bf.nbits
}

func (bf *Bitfield) Nbytes() int64 {
	return bf.nbytes
}

func (bf *Bitfield) Nset() int64 {
	return bf.nset
}
