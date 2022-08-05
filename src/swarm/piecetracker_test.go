package swarm

import (
	"testing"

	"gotor/bf"
)

func TestPeerPieceTracker_NextPieceByPeer(t *testing.T) {

	// Creates a bitfield from a slice of "have" indices
	bfFromHave := func(size int64, gots []int64) *bf.Bitfield {
		pbf := bf.NewBitfield(size)
		for _, idx := range gots {
			pbf.Set(idx, true)
		}
		return pbf
	}

	// peerAction is used to simulate a peer registering indices
	type peerAction struct {
		ph     *PeerHandler
		bf     []int64 // Which indices are set in the bitfield (if any)
		gets   []int64 // Which indices are set from "have messages"
		want   int64   // Desired result after all peerActions are processed
		leaves bool    // Should we de-register the peer after, Next() will not be tested on this peer
	}

	tests := []struct {
		name    string
		size    int64
		need    []int64 // What piece indices do we need
		actions []peerAction
	}{
		// Simple case
		{"bf_only", 10, []int64{0}, []peerAction{
			{&PeerHandler{}, []int64{0}, []int64{}, 0, false},
		}},

		{"have_only", 10, []int64{0}, []peerAction{
			{&PeerHandler{}, []int64{}, []int64{0}, 0, false},
		}},

		// Multiple peers
		{"multi_peer", 10, []int64{0, 1, 2}, []peerAction{
			{&PeerHandler{}, []int64{0, 1}, []int64{2}, 2, false},
			{&PeerHandler{}, []int64{0}, []int64{}, 0, false},
			{&PeerHandler{}, []int64{}, []int64{1}, 1, false},
		}},

		// Test leaving
		{"leaving_peers", 10, []int64{0, 1, 2}, []peerAction{
			{&PeerHandler{}, []int64{0, 2}, []int64{}, 2, false},
			{&PeerHandler{}, []int64{0}, []int64{}, 0, false},
			{&PeerHandler{}, []int64{2}, []int64{}, -1, true}, // want ignored
			{&PeerHandler{}, []int64{2}, []int64{}, -1, true}, // want ignored
		}},

		// A peer with no needed pieces
		{"no_good_pieces", 10, []int64{2}, []peerAction{
			{&PeerHandler{}, []int64{0, 1}, []int64{}, -1, false},
			{&PeerHandler{}, []int64{1, 2}, []int64{}, 2, false},
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ppt := NewPeerPieceTracker(tt.size)

			// Make bitfield
			bitfield := bfFromHave(tt.size, tt.need)

			// Simulate all the adding and deleting
			for _, action := range tt.actions {
				// Register BITFIELD message
				peerBf := bfFromHave(tt.size, action.bf)
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

				next := ppt.NextPieceByPeer(action.ph, bitfield)
				if next != action.want {
					t.Errorf("next piece for PeerHandler %p, got %v, want %v", action.ph, next, action.want)
				}
			}
		})
	}
}
