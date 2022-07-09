package peer

import (
	"fmt"
	"strings"
)

type List []Peer

// ListSource
// Any type that can extract a list of peers from itself.
type ListSource interface {
	GetPeers() (List, error)
}

func (l List) String() string {
	strb := strings.Builder{}
	strb.WriteString(fmt.Sprintf("Peer List (%v):", len(l)))
	for i, v := range l {
		if i%5 == 0 {
			strb.WriteString("\n\t")
		}
		strb.WriteString(fmt.Sprintf("(%v:%v) ", v.Ip(), v.Port()))
	}
	return strb.String()
}
