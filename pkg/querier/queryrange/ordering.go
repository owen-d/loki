package queryrange

import (
	"sort"

	"github.com/grafana/loki/pkg/logproto"
)

/*
Utils for manipulating ordering
*/

type marker []logproto.Entry

func (m marker) bounds() (int64, int64) {
	if len(m) == 0 {
		return 0, 0
	}
	return m[0].Timestamp.UnixNano(), m[len(m)-1].Timestamp.UnixNano()
}

type byDir struct {
	markers   []marker
	direction logproto.Direction
	labels    string
}

func (a byDir) Len() int      { return len(a.markers) }
func (a byDir) Swap(i, j int) { a.markers[i], a.markers[j] = a.markers[j], a.markers[i] }
func (a byDir) Less(i, j int) bool {
	x, _ := a.markers[i].bounds()
	y, _ := a.markers[j].bounds()

	if a.direction == logproto.BACKWARD {
		return x > y
	}
	return y > x
}
func (a byDir) merge() (result []logproto.Entry) {
	sort.Sort(a)
	for _, m := range a.markers {
		result = append(result, m...)
	}
	return result
}

// priorityqueue is used for extracting a limited # of entries from a set of sorted streams
type priorityqueue struct {
	streams   []*logproto.Stream
	direction logproto.Direction
}

func (pq *priorityqueue) Len() int { return len(pq.streams) }

func (pq *priorityqueue) Less(i, j int) bool {
	if pq.direction == logproto.FORWARD {
		return pq.streams[i].Entries[0].Timestamp.UnixNano() < pq.streams[j].Entries[0].Timestamp.UnixNano()
	}
	return pq.streams[i].Entries[0].Timestamp.UnixNano() > pq.streams[j].Entries[0].Timestamp.UnixNano()

}

func (pq *priorityqueue) Swap(i, j int) {
	pq.streams[i], pq.streams[j] = pq.streams[j], pq.streams[i]
}

func (pq *priorityqueue) Push(x interface{}) {
	stream := x.(*logproto.Stream)
	pq.streams = append(pq.streams, stream)
}

func (pq *priorityqueue) Pop() interface{} {
	n := pq.Len()
	stream := pq.streams[n-1]
	pq.streams[n-1] = nil // avoid memory leak
	pq.streams = pq.streams[:n-1]

	// put the rest of the stream back into the priorityqueue if more entries exist
	if len(stream.Entries) > 1 {
		remaining := *stream
		remaining.Entries = remaining.Entries[1:]
		pq.Push(&remaining)
	}

	stream.Entries = stream.Entries[:1]
	return stream
}
