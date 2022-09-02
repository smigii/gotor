package swarm

import (
	"net"
	"testing"

	"gotor/bf"
	"gotor/peer"
)

// Make an empty PeerHandler with random IP and port, and given ID
func PHDummy(id string) *PeerHandler {
	return &PeerHandler{
		peerInfo: peer.MakePeer(id, net.IP{}, 60666),
	}
}

// Creates a bitfield with all bits set to true, except those in the
// gots slice.
func bfFromNeed(size int64, gots []int64) *bf.Bitfield {
	pbf := bf.NewBitfield(size)
	// Set them all
	for i := int64(0); i < pbf.Nbits(); i++ {
		pbf.Set(i, true)
	}
	for _, idx := range gots {
		pbf.Set(idx, false)
	}
	return pbf
}

// Creates a bitfield with all bits set to false, except those in the
// gots slice.
func bfFromHave(size int64, gots []int64) *bf.Bitfield {
	pbf := bf.NewBitfield(size)
	for _, idx := range gots {
		pbf.Set(idx, true)
	}
	return pbf
}

func TestPeerPieceTracker_NextPiece(t *testing.T) {

	// peerAction is used to simulate a peer registering indices
	type peerAction struct {
		ph     *PeerHandler
		bf     []int64  // Which indices are set in the bitfield (if any)
		gets   []uint32 // Which indices are set from "have messages"
		want   uint32   // Desired result after all peerActions are processed
		succ   bool     // Should we recieve true or false
		leaves bool     // Should we de-register the peer after, Next() will not be tested on this peer
	}

	tests := []struct {
		name    string
		size    uint32
		need    []int64 // What piece indices do we need
		actions []peerAction
	}{
		// Simple case
		{"bf_only", 3, []int64{0}, []peerAction{
			{PHDummy("1"), []int64{0}, []uint32{}, 0, true, false},
		}},

		{"have_only", 3, []int64{0}, []peerAction{
			{PHDummy("1"), []int64{}, []uint32{0}, 0, true, false},
		}},

		// Multiple peers, rarest = 0 -> 1 -> 2
		{"multi_peer", 3, []int64{0, 1, 2}, []peerAction{
			{PHDummy("1"), []int64{0, 1, 2}, []uint32{}, 2, true, false},
			{PHDummy("2"), []int64{0, 1}, []uint32{}, 1, true, false},
			{PHDummy("3"), []int64{0}, []uint32{}, 0, true, false},
		}},

		// Test leaving
		{"leaving_peers", 2, []int64{0, 1}, []peerAction{
			{PHDummy("1"), []int64{0, 1}, []uint32{}, 1, true, false},
			{PHDummy("2"), []int64{0}, []uint32{}, 0, true, false},
			{PHDummy("3"), []int64{1}, []uint32{}, 00, false, true}, // want ignored
			{PHDummy("4"), []int64{1}, []uint32{}, 00, false, true}, // want ignored
		}},

		// A peer with no needed pieces
		{"no_good_pieces", 10, []int64{2}, []peerAction{
			{PHDummy("1"), []int64{0, 1}, []uint32{}, 00, false, false},
			{PHDummy("2"), []int64{1, 2}, []uint32{}, 2, true, false},
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Make bitfield
			bitfield := bfFromNeed(int64(tt.size), tt.need)

			ppt := NewPeerPieceTracker(tt.size, bitfield)

			// Simulate all the adding and deleting
			for _, action := range tt.actions {
				// Register BITFIELD message
				peerBf := bfFromHave(int64(tt.size), action.bf)
				ppt.RegisterBF(action.ph, peerBf)

				// Register HAVE messages
				ppt.Register(action.ph, action.gets...)

				// Peer disconnects
				if action.leaves {
					ppt.Unregister(action.ph)
				}
			}

			for _, action := range tt.actions {
				// Skip peers that unregistered
				if action.leaves {
					continue
				}

				next, ok := ppt.NextPiece(action.ph)
				if action.succ && !ok {
					t.Errorf("next piece for [%v], no next piece found", action.ph.Key())
				}
				if next != action.want {
					t.Errorf("next piece for [%v], got %v, want %v", action.ph.Key(), next, action.want)
				}
			}
		})
	}
}
