package tracker

// Stats to be sent in a tracker GET request. Uploaded and Dnloaded
// should report total bytes up/dnloaded, while Left should report
// how many bytes until we have 100% downloaded.
type Stats struct {
	uploaded uint64
	dnloaded uint64
	left     uint64
}

func NewStats(dn uint64, up uint64, left uint64) *Stats {
	return &Stats{
		uploaded: up,
		dnloaded: dn,
		left:     left,
	}
}

func (s *Stats) IncUploaded(amnt uint64) {
	s.uploaded += amnt
}

func (s *Stats) IncDnloaded(amnt uint64) {
	s.dnloaded += amnt
}

func (s *Stats) DecLeft(amnt uint64) {
	if amnt >= s.left {
		s.left = 0
	} else {
		s.left -= amnt
	}
}

func (s Stats) Uploaded() uint64 {
	return s.uploaded
}

func (s Stats) Dnloaded() uint64 {
	return s.dnloaded
}

func (s Stats) Left() uint64 {
	return s.left
}
