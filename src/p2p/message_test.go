package p2p

import (
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
