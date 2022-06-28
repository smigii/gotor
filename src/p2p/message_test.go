package p2p

import (
	"bytes"
	"testing"
)

func TestDecodeAll(t *testing.T) {
	tests := []struct {
		name   string
		data   []byte
		types  []uint8 // Types of encoded messages
		toRead uint64  // How much should be read by DecodeAll
	}{
		{
			"Good data",
			[]byte{
				0x0, 0x0, 0x0, 0x0, // Keep Alive
				0x0, 0x0, 0x0, 0x1, 0x0, // Choke
				0x0, 0x0, 0x0, 0x1, 0x1, // Unchoke
				0x0, 0x0, 0x0, 0x5, 0x4, 0x0, 0x0, 0x0, 0xf, // Have
			},
			[]uint8{TypeKeepAlive, TypeChoke, TypeUnchoke, TypeHave},
			23, // Length of data
		},
		{
			"Unknown type",
			[]byte{
				0x0, 0x0, 0x0, 0x0, // Keep Alive
				0x0, 0x0, 0x0, 0x1, 0x0, // Choke
				0x0, 0x0, 0x0, 0x1, 0x1, // Unchoke
				0x0, 0x0, 0x0, 0x1, 0xDD, // BAD TYPE
				0x0, 0x0, 0x0, 0x5, 0x4, 0x0, 0x0, 0x0, 0xf, // Have
			},
			[]uint8{TypeKeepAlive, TypeChoke, TypeUnchoke},
			14, // Keep Alive + Choke + Unchoke
		},
		{
			"Bad HAVE",
			[]byte{
				0x0, 0x0, 0x0, 0x0, // Keep Alive
				0x0, 0x0, 0x0, 0x1, 0x0, // Choke
				0x0, 0x0, 0x0, 0x1, 0x1, // Unchoke
				0x0, 0x0, 0x0, 0x7, 0x4, 0x0, 0x0, 0x0, 0xf, // BAD HAVE
				0x0, 0x0, 0x0, 0x5, 0x4, 0x0, 0x0, 0x0, 0xf, // Have
				0x0, 0x0, 0x0, 0x0, // Keep Alive
			},
			[]uint8{TypeKeepAlive, TypeChoke, TypeUnchoke, TypeHave, TypeKeepAlive},
			27, // All - bad have
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dar := DecodeAll(tt.data)

			// How many messages were decoded
			if len(dar.Msgs) != len(tt.types) {
				t.Errorf("Decoded %v messages, expected %v", len(dar.Msgs), len(tt.types))
			}

			// Check amount read
			if dar.Read != tt.toRead {
				t.Errorf("Read %v bytes, should have read %v bytes", dar.Read, tt.toRead)
			}

			// Check correct types
			for i, msg := range dar.Msgs {
				if msg.Mtype() != tt.types[i] {
					t.Errorf("Type of message %v should be %v, got %v", i, tt.types[i], msg.Mtype())
				}
			}
		})
	}
}

// This simulates a conn.recv() loop, where the buffer being read into is being
// overwritten each loop. Need to make sure the bitfields are not also being
// overwritten.
func TestBitfieldMem(t *testing.T) {
	tests := []struct {
		name string
		bfs  [][]byte
		bfs2 [][]byte
	}{
		{
			name: "",
			bfs: [][]byte{
				{0, 0, 0, 2, 5, 20},
				{0, 0, 0, 2, 5, 40},
			},
			bfs2: [][]byte{
				{0, 0, 0, 2, 5, 50},
				{0, 0, 0, 2, 5, 70},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// This is the buffer we would recv() into
			buf := make([]byte, 4096, 4096)

			// Build a raw message
			data := make([]byte, 0, 16)
			for _, bf := range tt.bfs {
				data = append(data, bf...)
			}
			// Simulate the recv() call
			copy(buf, data)

			// Get the first batch of messages
			dar := DecodeAll(buf[:len(data)])
			msgs := dar.Msgs

			// Build another raw message
			data2 := make([]byte, 0, 16)
			for _, bf := range tt.bfs2 {
				data2 = append(data2, bf...)
			}
			// Simulate the recv() call
			copy(buf, data2)

			// Get the first batch of messages
			dar2 := DecodeAll(buf[:len(data2)])
			msgs2 := dar2.Msgs

			// Check that the first batch of messages haven't been overwritten
			if len(msgs) != len(tt.bfs) {
				t.Errorf("bad decode all lengths 1\nwant: %v\n got: %v", len(msgs), len(tt.bfs))
			}
			for i, bf := range tt.bfs {
				got := msgs[i].Encode()
				if !bytes.Equal(bf, got) {
					t.Errorf("bad decode all 1\nwant: %v\n got: %v", bf, got)
				}
			}

			// Check that the second batch of messages hasn't been overwritten
			if len(msgs2) != len(tt.bfs2) {
				t.Errorf("bad decode all lengths 2\nwant: %v\n got: %v", len(msgs2), len(tt.bfs2))
			}
			for i, bf := range tt.bfs2 {
				got := msgs2[i].Encode()
				if !bytes.Equal(bf, got) {
					t.Errorf("bad decode all 2\nwant: %v\n got: %v", bf, got)
				}
			}

		})
	}
}
