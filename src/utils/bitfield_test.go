package utils

import "testing"

func TestGet(t *testing.T) {
	bytes := []byte{0xAA, 0xF0}
	bf := FromBytes(bytes, 16)

	tests := []struct {
		idx  uint64
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
	bf := FromBytes(bytes, 16)

	tests := []struct {
		idx uint64
		to  bool
	}{
		{0, true},
		{2, true},
		{8, false},
		{10, false},
	}
	for _, tt := range tests {
		bf.Set(tt.idx, tt.to)
		got := bf.Get(tt.idx)
		if got != tt.to {
			t.Errorf("Get() = %v, want %v", got, tt.to)
		}
	}
}
