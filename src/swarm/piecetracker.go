package swarm

import (
	"sync"

	"gotor/bf"
	"gotor/utils/ll"
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

// PeerPieceTracker tracks which peers have which pieces, and provides a fast
// lookup for the rarest pieces by peer.
type PeerPieceTracker struct {
	nodes []*ll.Node[piece] // Orders pieces by index

	// Orders pieces by count. Each index i corresponds to the list of pieces
	// that have count i. At the start of a swarm, all pieces should be in
	// the list at index 0, since all pieces start with 0 count.
	buckets []ll.LinkedList[piece]

	mutex sync.Mutex
}

type piece struct {
	index int64
	peers map[*PeerHandler]struct{} // Set of peers who have this piece
}

// ============================================================================
// ============================================================================

func NewPeerPieceTracker(size int64) *PeerPieceTracker {
	ppt := PeerPieceTracker{}

	ppt.nodes = make([]*ll.Node[piece], size, size)
	ppt.buckets = make([]ll.LinkedList[piece], numBuckets, numBuckets)

	// Initialize nodes
	for i := int64(0); i < size; i++ {
		p := piece{
			index: i,
			peers: make(map[*PeerHandler]struct{}),
		}
		node := ppt.buckets[0].AddDataFront(p)
		ppt.nodes[i] = node
	}

	return &ppt
}

// Register registers the given peer as having the given piece indices.
func (ppt *PeerPieceTracker) Register(whom *PeerHandler, indices ...int64) {
	ppt.mutex.Lock()
	defer ppt.mutex.Unlock()

	for _, index := range indices {
		node := ppt.nodes[index]

		// Update piece's "peers" set
		if _, has := node.Data.peers[whom]; !has {
			// Move to next bucket, unless this is the largest bucket
			oldCount := len(node.Data.peers)
			if oldCount < numBuckets {
				ppt.buckets[oldCount].Remove(node)
				ppt.buckets[oldCount+1].AddNodeFront(node)
			}

			// Add to set
			node.Data.peers[whom] = struct{}{}
		}
		//node.Data.count++
	}
}

// Unregister removes the given peer from all piece's peer sets.
func (ppt *PeerPieceTracker) Unregister(whom *PeerHandler) {
	ppt.mutex.Lock()
	defer ppt.mutex.Unlock()

	for _, node := range ppt.nodes {

		if _, has := node.Data.peers[whom]; has {
			// Move to previous bucket, unless decrementing value by 1
			// leaves it in the largest bucket. I.e, numBuckets = 50
			// and oldCount = 52, then the node should remain in bucket 50.
			oldCount := len(node.Data.peers)
			if oldCount <= numBuckets {
				ppt.buckets[oldCount].Remove(node)
				ppt.buckets[oldCount-1].AddNodeFront(node)
			}

			// Remove from set
			delete(node.Data.peers, whom)
		}
	}
}

// NextPieceByPeer returns the rarest piece that has been registered to the
// given peer, and that isn't set in the given bitfield. Pieces with count 0
// are skipped. If the peer hasn't registered any pieces that are needed,
// returns -1.
func (ppt *PeerPieceTracker) NextPieceByPeer(whom *PeerHandler, need *bf.Bitfield) int64 {
	ppt.mutex.Lock()
	defer ppt.mutex.Unlock()

	for i := 1; i < len(ppt.buckets); i++ {

		cur := ppt.buckets[i].Head()
		for cur != nil {
			// If the piece is needed
			if need.Get(cur.Data.index) {
				// If the peer has the piece
				if _, has := cur.Data.peers[whom]; has {
					return cur.Data.index
				}
			}
			cur = cur.Next()
		}
	}

	return -1
}
