package info

import (
	"fmt"

	"gotor/torrent/filesd"
)

type pieceLocation struct {
	Entry *filesd.Entry
	Loc   filesd.PieceInfo
}

// PieceMap is contains numPieces slices of pieceLocations, lets us map
// piece indices to a slice of pieceLocations.
type PieceMap [][]pieceLocation

func MakePieceMap(fl filesd.FileList, npieces int64, pieceLen int64, totalLen int64) (PieceMap, error) {

	pm := make(PieceMap, npieces, npieces)
	curFileIdx := int64(0)
	curEntry := &fl[curFileIdx]
	nfiles := int64(len(fl))
	pieceRem := pieceLen // How much of the current piece (i) is unaccounted for
	totalCounted := int64(0)

	var i int64
	for i = 0; i < npieces; i++ {

		for {

			if pieceRem > 0 {

				pinfo, e := curEntry.PieceInfo(i, pieceLen)
				if e != nil {
					// Remainder of piece must be in next file
					curFileIdx++
					if curFileIdx == nfiles {
						if totalCounted == totalLen {
							// Last piece may be truncated
							break
						} else {
							return nil, fmt.Errorf("no file contained piece %v", i)
						}
					}
					curEntry = &fl[curFileIdx]
					continue
				} else {
					pieceRem -= pinfo.ReadAmnt
					totalCounted += pinfo.ReadAmnt

					// Append the piece info to the piece map
					pieceLoc := pieceLocation{
						Entry: curEntry,
						Loc:   pinfo,
					}
					pm[i] = append(pm[i], pieceLoc)

					if pinfo.ReadAmnt != pieceLen {
						// File must be fully read
						curFileIdx++
						if curFileIdx == nfiles {
							if totalCounted == totalLen {
								// Last piece may be truncated
								break
							} else {
								return nil, fmt.Errorf("no file contained piece %v", i)
							}
						}
						curEntry = &fl[curFileIdx]
						continue
					}
				}

			} else {
				// Current piece (i) is fully accounted for, start
				// processing on next piece
				pieceRem = pieceLen
				break
			}

		}

	}

	return pm, nil
}

type pieceMapStateMachine struct {
	pm           PieceMap
	flist        filesd.FileList
	curEntry     *filesd.Entry
	curFileIdx   int64
	nfiles       int64
	pieceRem     int64 // How much of the current piece (i) is unaccounted for
	totalCounted int64
}

func makePieceMapStateMachine(npieces int64, flist filesd.FileList, pieceLen int64) pieceMapStateMachine {
	return pieceMapStateMachine{
		pm:           make(PieceMap, npieces, npieces),
		flist:        flist,
		curFileIdx:   0,
		curEntry:     &flist[0],
		nfiles:       int64(len(flist)),
		pieceRem:     pieceLen,
		totalCounted: 0,
	}
}

func MakePieceMap2(fl filesd.FileList, npieces int64, pieceLen int64, totalLen int64) (PieceMap, error) {

	pm := make(PieceMap, npieces, npieces)

	pieceIdx := int64(0)
	pieceRem := pieceLen
	totalCounted := int64(0)

	for fidx, fe := range fl {

		if pieceIdx == npieces {
			break
		}

		// Keep getting pieces for the file until we get a nil. Then,
		// process the next file entry.
		for {
			pinfo, e := fe.PieceInfo(pieceIdx, pieceLen)
			if e != nil {
				// Start looking through next file entry
				break
			} else {
				pLoc := pieceLocation{
					Entry: &fl[fidx],
					Loc:   pinfo,
				}
				pm[pieceIdx] = append(pm[pieceIdx], pLoc)

				pieceRem -= pinfo.ReadAmnt
				totalCounted += pinfo.ReadAmnt

				if pieceRem <= 0 {
					pieceRem = pieceLen
					pieceIdx++
				} else {
					// If there is still more of the piece to be read, it must
					// in the next file
					break
				}
			}
		}

	}

	if totalCounted != totalLen {
		return nil, fmt.Errorf("only accounted for %v/%v bytes", totalCounted, totalLen)
	}

	return pm, nil
}
