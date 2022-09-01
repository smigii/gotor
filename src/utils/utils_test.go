package utils

import (
	"reflect"
	"testing"
)

func TestSegmentData(t *testing.T) {
	tests := []struct {
		name    string
		data    []byte
		segSize int64
		want    [][]byte
	}{
		{"Single", []byte{0, 1, 2, 3}, 4, [][]byte{{0, 1, 2, 3}}},
		{"Exact", []byte{0, 1, 2, 3, 4, 5}, 2, [][]byte{{0, 1}, {2, 3}, {4, 5}}},
		{"Smaller Last", []byte{0, 1, 2, 3, 4, 5, 6, 7}, 3, [][]byte{{0, 1, 2}, {3, 4, 5}, {6, 7}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := SegmentData(tt.data, tt.segSize)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("SegmentData() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRemoveSwap(t *testing.T) {
	tests := []struct {
		name string
		s    []int
		rm   int64
		want []int
	}{ // yikes test but whatever
		{"", []int{0, 1, 2, 3, 4}, 0, []int{4, 1, 2, 3}},
		{"", []int{0, 1, 2, 3, 4}, 1, []int{0, 4, 2, 3}},
		{"", []int{0, 1, 2, 3, 4}, 2, []int{0, 1, 4, 3}},
		{"", []int{0, 1, 2, 3, 4}, 3, []int{0, 1, 2, 4}},
		{"", []int{0, 1, 2, 3, 4}, 4, []int{0, 1, 2, 3}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.s = RemoveSwap[int](tt.s, tt.rm)
			if !reflect.DeepEqual(tt.s, tt.want) {
				t.Errorf(" got %v\nwant %v", tt.s, tt.want)
			}
		})
	}
}
