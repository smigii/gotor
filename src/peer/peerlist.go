package peer

type List []*Peer

// ListSource
// Any type that can extract a list of peers from itself.
type ListSource interface {
	GetPeers() (List, error)
}
