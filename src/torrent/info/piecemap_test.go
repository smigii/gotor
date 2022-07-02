package info

import (
	"testing"

	"gotor/torrent/filesd"
	"gotor/utils/test"
)

func TestMakePieceMap(t *testing.T) {

	type testLocation struct {
		torPath  string
		seekAmnt int64
		readAmnt int64
	}

	tests := []struct {
		name     string
		pieceLen int64
		totalLen int64
		fentries []filesd.EntryBase
		wantLoc  [][]testLocation
	}{
		// [A|A|A|A|A]
		{"", 5, 5, []filesd.EntryBase{
			filesd.MakeFileEntry("A", 5),
		},
			[][]testLocation{
				{{"A", 0, 5}}, // Piece 0
			},
		},

		// [A|A|A| | ]
		{"", 5, 3, []filesd.EntryBase{
			filesd.MakeFileEntry("A", 3),
		},
			[][]testLocation{
				{{"A", 0, 3}}, // Piece 0
			},
		},

		// [A|A|A] [A|A|A] [A|A| ]
		{"", 3, 8, []filesd.EntryBase{
			filesd.MakeFileEntry("A", 8),
		},
			[][]testLocation{
				{{"A", 0, 3}}, // Piece 0
				{{"A", 3, 3}}, // Piece 1
				{{"A", 6, 2}}, // Piece 2
			},
		},

		// [A|A|A]  [B|B|B]  [C|C|C]
		{"", 3, 9, []filesd.EntryBase{
			filesd.MakeFileEntry("A", 3),
			filesd.MakeFileEntry("B", 3),
			filesd.MakeFileEntry("C", 3),
		},
			[][]testLocation{
				{{"A", 0, 3}}, // Piece 0
				{{"B", 0, 3}}, // Piece 1
				{{"C", 0, 3}}, // Piece 2
			},
		},

		// [A|A|A]  [A|B|B]  [B|B|B]  [B|C|C]
		{"", 3, 12, []filesd.EntryBase{
			filesd.MakeFileEntry("A", 4),
			filesd.MakeFileEntry("B", 6),
			filesd.MakeFileEntry("C", 2),
		},
			[][]testLocation{
				{{"A", 0, 3}},              // Piece 0
				{{"A", 3, 1}, {"B", 0, 2}}, // Piece 1
				{{"B", 2, 3}},              // Piece 2
				{{"B", 5, 1}, {"C", 0, 2}}, // Piece 3
			},
		},

		// [A|A|A|A]  [A|B|C|D]  [E| | | ]
		{"", 4, 9, []filesd.EntryBase{
			filesd.MakeFileEntry("A", 5),
			filesd.MakeFileEntry("B", 1),
			filesd.MakeFileEntry("C", 1),
			filesd.MakeFileEntry("D", 1),
			filesd.MakeFileEntry("E", 1),
		},
			[][]testLocation{
				{{"A", 0, 4}},
				{{"A", 4, 1}, {"B", 0, 1}, {"C", 0, 1}, {"D", 0, 1}},
				{{"E", 0, 1}},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			npieces := filesd.CalcNumPieces(tt.fentries, tt.pieceLen)
			flist := filesd.MakeFileList(tt.fentries, tt.pieceLen)
			pm, e := MakePieceMap2(flist, npieces, tt.pieceLen, tt.totalLen)
			test.CheckError(t, e)

			// For each piece
			for idx, pieceLocSlice := range pm {

				if len(pieceLocSlice) != len(tt.wantLoc[idx]) {
					t.Errorf("index %v mismatched lens, got %v, want %v", idx, len(pieceLocSlice), len(tt.wantLoc[idx]))
				}

				// Check we have the same slice of piece locations
				for j, ploc := range pieceLocSlice {

					wantLocation := tt.wantLoc[idx][j]

					if ploc.Entry.TorPath() != wantLocation.torPath {
						t.Errorf("index [%v,%v] got path %v, want %v", idx, j, ploc.Entry.TorPath(), wantLocation.torPath)
					}

					if ploc.Loc.SeekAmnt != wantLocation.seekAmnt {
						t.Errorf("index [%v,%v] got seekAmnt %v, want %v", idx, j, ploc.Loc.SeekAmnt, wantLocation.seekAmnt)
					}

					if ploc.Loc.ReadAmnt != wantLocation.readAmnt {
						t.Errorf("index [%v,%v] got readAmnt %v, want %v", idx, j, ploc.Loc.ReadAmnt, wantLocation.readAmnt)
					}

				}

			}

		})
	}
}
