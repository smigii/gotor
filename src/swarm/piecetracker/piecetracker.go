package piecetracker

import (
	"sync"

	"gotor/utils/linklist"
)

// ============================================================================
// ============================================================================

const (
	// numBuckets is the number of piece count buckets. We should really only
	// have a maximum piece count of 50, since we will only be holding 50
	// connections max at a time. If a piece has over numBuckets availablity, it
	// really doesn't need any finer-grained sorting since that's already very
	// high, and it can just stay in the highest count bucket
	numBuckets = 64
)

// ============================================================================
// ============================================================================

type PeerPieceTracker struct {
	nodes   []*linklist.Node[piece]      // Orders pieces by index
	buckets []linklist.LinkedList[piece] // Orders pieces by count
	lock    sync.Mutex
}

type piece struct {
	index int64
	count int16
}

// ============================================================================
// ============================================================================

func NewPeerPieceTracker(size int64) *PeerPieceTracker {
	ppt := PeerPieceTracker{}

	ppt.nodes = make([]*linklist.Node[piece], size, size)
	ppt.buckets = make([]linklist.LinkedList[piece], numBuckets, numBuckets)

	// Initialize nodes
	for i := int64(0); i < size; i++ {
		p := piece{
			index: i,
			count: 0,
		}
		node := ppt.buckets[0].AddDataFront(p)
		ppt.nodes[i] = node
	}

	return &ppt
}

// NextPiece returns the piece index with the lowest count that also exists in
// the provided bitfield. Pieces with count 0 are excluded. If no piece is
// found, returns 0.
func (ppt *PeerPieceTracker) NextPiece(indices []bool) int64 {
	ppt.lock.Lock()
	defer ppt.lock.Unlock()

	for i := 1; i < len(ppt.buckets); i++ {

		cur := ppt.buckets[i].Head()
		for cur != nil {
			if indices[cur.Data.index] {
				return cur.Data.index
			}
			cur = cur.Next()
		}
	}

	return -1
}

// IncPieces increments the piece counts for the given indices.
func (ppt *PeerPieceTracker) IncPieces(indices ...int64) {
	ppt.lock.Lock()
	defer ppt.lock.Unlock()

	for _, index := range indices {
		node := ppt.nodes[index]
		oldCount := node.Data.count
		ppt.buckets[oldCount].Remove(node)
		ppt.buckets[oldCount+1].AddNodeFront(node)
		node.Data.count++
	}
}

// DecPieces decrements the piece counts for the given indices.
func (ppt *PeerPieceTracker) DecPieces(indices ...int64) {
	ppt.lock.Lock()
	defer ppt.lock.Unlock()

	for _, index := range indices {
		node := ppt.nodes[index]
		oldCount := node.Data.count
		ppt.buckets[oldCount].Remove(node)
		ppt.buckets[oldCount-1].AddNodeFront(node)
		node.Data.count--
	}
}
