package swarm

import (
	"gotor/utils/ds"
	"sync"

	"gotor/bf"
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
	nodes []*ds.Node[piece] // Orders pieces by index

	// Orders pieces by count. Each index i corresponds to the list of pieces
	// that have count i. At the start of a swarm, all pieces should be in
	// the list at index 0, since all pieces start with 0 count.
	buckets []ds.LinkedList[piece]

	// Maps which peers are downloading which pieces
	requests map[*PeerHandler][]*piece

	bf *bf.Bitfield // Our bitfield

	mutex sync.Mutex
}

type piece struct {
	index   uint32
	active  bool // Is this index being requested from a peer?
	peerSet ds.Set[*PeerHandler]
}

// ============================================================================
// ============================================================================

func NewPeerPieceTracker(size uint32, bf *bf.Bitfield) *PeerPieceTracker {
	ppt := PeerPieceTracker{}

	ppt.nodes = make([]*ds.Node[piece], size, size)
	ppt.buckets = make([]ds.LinkedList[piece], numBuckets, numBuckets)
	ppt.requests = make(map[*PeerHandler][]*piece)
	ppt.bf = bf

	// Initialize nodes
	for i := uint32(0); i < size; i++ {
		p := piece{
			index:   i,
			active:  false,
			peerSet: ds.MakeSet[*PeerHandler](),
		}
		node := ppt.buckets[0].AddDataFront(p)
		ppt.nodes[i] = node
	}

	return &ppt
}

// RegisterBF registers all the set bits of a bitfield to the fiven peer.
func (ppt *PeerPieceTracker) RegisterBF(whom *PeerHandler, bf *bf.Bitfield) {
	ppt.mutex.Lock()
	defer ppt.mutex.Unlock()

	for index := int64(0); index < bf.Nbits(); index++ {
		if bf.Get(index) {
			ppt.register(whom, uint32(index))
		}
	}
}

// Register registers the given peer as having the given piece indices.
func (ppt *PeerPieceTracker) Register(whom *PeerHandler, indices ...uint32) {
	ppt.mutex.Lock()
	defer ppt.mutex.Unlock()

	for _, index := range indices {
		ppt.register(whom, index)
	}
}

// register will register the given peer as having the given piece indices.
// This should only be called by Register or RegisterBF, which have acquired
// the lock.
func (ppt *PeerPieceTracker) register(whom *PeerHandler, index uint32) {
	node := ppt.nodes[index]

	// Update piece's "peers" set
	if !node.Data.peerSet.Has(whom) {
		// Move to next bucket, unless this is the largest bucket
		oldCount := node.Data.peerSet.Size()
		if oldCount < numBuckets {
			ppt.buckets[oldCount].Remove(node)
			ppt.buckets[oldCount+1].AddNodeFront(node)
		}

		// Add to set
		node.Data.peerSet.Add(whom)
	}
}

// Unregister removes the given peer from all piece's peer sets.
func (ppt *PeerPieceTracker) Unregister(whom *PeerHandler) {
	ppt.mutex.Lock()
	defer ppt.mutex.Unlock()

	// Set the peers active requests to unactive
	for _, p := range ppt.requests[whom] {
		p.active = false
	}

	// Remove peer from all index peer sets
	for _, node := range ppt.nodes {

		if node.Data.peerSet.Has(whom) {
			// Move to previous bucket, unless decrementing value by 1
			// leaves it in the largest bucket. I.e, numBuckets = 50
			// and oldCount = 52, then the node should remain in bucket 50.
			oldCount := node.Data.peerSet.Size()
			if oldCount <= numBuckets {
				ppt.buckets[oldCount].Remove(node)
				ppt.buckets[oldCount-1].AddNodeFront(node)
			}

			// Remove from set
			node.Data.peerSet.Remove(whom)
		}
	}
}

// NextPiece gets the rarest piece index that is available to download
// from given PeerHandler, and that is not being downloaded by any other
// peer. The returned index will be marked as active, and no other peer
// may acquire it. If no piece index is available, returns (0, false)
func (ppt *PeerPieceTracker) NextPiece(whom *PeerHandler) (uint32, bool) {
	ppt.mutex.Lock()
	defer ppt.mutex.Unlock()

	for i := 1; i < len(ppt.buckets); i++ {

		cur := ppt.buckets[i].Head()
		for cur != nil {
			curPiece := cur.Data
			// If the piece is needed
			if ppt.bf.Get(int64(curPiece.index)) {
				// If the piece isn't taken by another peer, and
				// the peer has the piece
				if !curPiece.active && curPiece.peerSet.Has(whom) {
					curPiece.active = true
					ppt.requests[whom] = append(ppt.requests[whom], &curPiece)
					return curPiece.index, true
				}
			}
			cur = cur.Next()
		}
	}

	return 0, false
}
