package utils

import (
	"testing"
)

func TestFromBytes(t *testing.T) {

	tests := []struct {
		name  string
		data  []byte
		nbits int64
		nset  int64
		err   bool
	}{
		{"1 byte 8 bit", []byte{0x0A}, 8, 2, false},
		{"1 byte 4 bit", []byte{0xFF}, 4, 4, false},
		{"1 byte 9 bit", []byte{0x00}, 9, 0, true},
		{"2 byte 16 bit", []byte{0xAA, 0xAA}, 16, 8, false},
		{"2 byte 9 bit", []byte{0xFF, 0xFF}, 9, 9, false},
		{"2 byte 13 bit", []byte{0xFF, 0x8}, 13, 9, false},
		{"2 byte 8 bit", []byte{0x00, 0x00}, 8, 0, true},
		{"2 byte 17 bit", []byte{0x00, 0x00}, 17, 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			bf, e := FromBytes(tt.data, tt.nbits)
			if tt.err {
				if e == nil {
					t.Errorf("expected error, got no error")
				}
			} else {
				if e != nil {
					t.Error(e)
				}
				if bf.nset != tt.nset {
					t.Errorf("expected nset to be %v, got %v", tt.nset, bf.nset)
				}
			}

		})
	}
}

func TestGet(t *testing.T) {
	bytes := []byte{0xAA, 0xF0}
	bf, e := FromBytes(bytes, 16)
	if e != nil {
		t.Fatal(e)
	}

	tests := []struct {
		idx  int64
		want bool
	}{
		{0, true},
		{1, false},
		{2, true},
		{3, false},
		{8, true},
		{9, true},
		{10, true},
		{11, true},
	}
	for _, tt := range tests {
		got := bf.Get(tt.idx)
		if got != tt.want {
			t.Errorf("Get() = %v, want %v", got, tt.want)
		}
	}
}

func TestSet(t *testing.T) {
	bytes := []byte{0x00, 0xFF}
	bf, e := FromBytes(bytes, 16)
	if e != nil {
		t.Fatal(e)
	}

	tests := []struct {
		idx   int64
		to    bool
		delta int // Change to Bitfield.nset
	}{
		{0, true, 1},
		{0, true, 0},
		{2, true, 1},
		{2, true, 0},
		{8, false, -1},
		{8, false, 0},
		{10, false, -1},
		{10, false, 0},
	}
	for _, tt := range tests {
		prevNset := int(bf.nset)

		bf.Set(tt.idx, tt.to)
		got := bf.Get(tt.idx)
		if got != tt.to {
			t.Errorf("Get() = %v, want %v", got, tt.to)
		}
		if (int(bf.nset) - prevNset) != tt.delta {
			t.Errorf("expected nset to change by %v, prev = %v, new = %v", tt.delta, prevNset, bf.nset)
		}
	}
}
