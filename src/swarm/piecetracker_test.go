package swarm

import (
	"testing"

	"gotor/bf"
)

func TestPeerPieceTracker_NextPieceByPeer(t *testing.T) {

	// peerAction is used to simulate a peer registering indices
	type peerAction struct {
		ph     *PeerHandler
		has    []int64
		want   int64 // Desired result after all peerActions are processed
		leaves bool  // Should we de-register the peer after, Next() will not be tested on this peer
	}

	tests := []struct {
		name    string
		size    int64
		need    []int64 // What piece indices do we need
		actions []peerAction
	}{
		// Simple case
		{"", 10, []int64{0}, []peerAction{
			{&PeerHandler{}, []int64{0}, 0, false},
		}},

		// Multiple peers
		{"", 10, []int64{0, 1, 2}, []peerAction{
			{&PeerHandler{}, []int64{0, 1, 2}, 2, false},
			{&PeerHandler{}, []int64{0}, 0, false},
			{&PeerHandler{}, []int64{1}, 1, false},
		}},

		// Test leaving
		{"", 10, []int64{0, 1, 2}, []peerAction{
			{&PeerHandler{}, []int64{0, 2}, 2, false},
			{&PeerHandler{}, []int64{0}, 0, false},
			{&PeerHandler{}, []int64{2}, -1, true}, // want ignored
			{&PeerHandler{}, []int64{2}, -1, true}, // want ignored
		}},

		// A peer with no needed pieces
		{"", 10, []int64{2}, []peerAction{
			{&PeerHandler{}, []int64{0, 1}, -1, false},
			{&PeerHandler{}, []int64{1, 2}, 2, false},
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ppt := NewPeerPieceTracker(tt.size)

			// Make bitfield
			bools := make([]bool, tt.size, tt.size)
			for _, i := range tt.need {
				bools[i] = true
			}
			bitfield := bf.FromBoolSlice(bools)

			// Simulate all the adding and deleting
			for _, action := range tt.actions {
				ppt.Register(action.ph, action.has...)
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
