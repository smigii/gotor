package peer

type State struct {
	chokingUs    bool // Peer is choking us
	weChoking    bool // We are choking peer
	interestedUs bool // Peer is interested in us
	weInterested bool // We are interested in peer
}

func MakeState() State {
	return State{
		chokingUs:    true,
		weChoking:    true,
		interestedUs: false,
		weInterested: false,
	}
}

func (s *State) ChokingUs() bool {
	return s.chokingUs
}

func (s *State) SetChokingUs(chokingUs bool) {
	s.chokingUs = chokingUs
}

func (s *State) WeChoking() bool {
	return s.weChoking
}

func (s *State) SetWeChoking(weChoking bool) {
	s.weChoking = weChoking
}

func (s *State) InterestedUs() bool {
	return s.interestedUs
}

func (s *State) SetInterestedUs(interestedUs bool) {
	s.interestedUs = interestedUs
}

func (s *State) WeInterested() bool {
	return s.weInterested
}

func (s *State) SetWeInterested(weInterested bool) {
	s.weInterested = weInterested
}
