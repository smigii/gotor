package info

import (
	"fmt"

	"gotor/torrent/filesd"
)

// PieceMap is contains numPieces slices of pieceLocations, lets us map
// piece indices to a slice of pieceLocations.
type PieceMap [][]PieceLocation

func MakePieceMap(fl filesd.FileList, npieces int64, pieceLen int64, totalLen int64) (PieceMap, error) {

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
				pLoc := PieceLocation{
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
