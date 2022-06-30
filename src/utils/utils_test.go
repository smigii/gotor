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
