package queryrange

import (
	"sort"

	"github.com/grafana/loki/pkg/logproto"
)

/*
Utils for manipulating ordering
*/

type marker []logproto.Entry

func (m marker) start() int64 {
	if len(m) == 0 {
		return 0
	}
	return m[0].Timestamp.UnixNano()
}

type byDir struct {
	markers   []marker
	direction logproto.Direction
	labels    string
}

func (a byDir) Len() int      { return len(a.markers) }
func (a byDir) Swap(i, j int) { a.markers[i], a.markers[j] = a.markers[j], a.markers[i] }
func (a byDir) Less(i, j int) bool {
	x, y := a.markers[i].start(), a.markers[j].start()

	if a.direction == logproto.BACKWARD {
		return x > y
	}
	return y > x
}
func (a byDir) EntriesCount() (n int) {
	for _, m := range a.markers {
		n += len(m)
	}
	return n
}

func (a byDir) merge() []logproto.Entry {
	result := make([]logproto.Entry, 0, a.EntriesCount())

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

// ---------------------------------- old/unused below, but used for benchmarking ---------------------------------------------
type entry struct {
	entry  logproto.Entry
	labels string
}

type byDirection struct {
	direction logproto.Direction
	entries   []entry
}

func (a byDirection) Len() int      { return len(a.entries) }
func (a byDirection) Swap(i, j int) { a.entries[i], a.entries[j] = a.entries[j], a.entries[i] }
func (a byDirection) Less(i, j int) bool {
	e1, e2 := a.entries[i], a.entries[j]
	if a.direction == logproto.BACKWARD {
		switch {
		case e1.entry.Timestamp.UnixNano() < e2.entry.Timestamp.UnixNano():
			return false
		case e1.entry.Timestamp.UnixNano() > e2.entry.Timestamp.UnixNano():
			return true
		default:
			return e1.labels > e2.labels
		}
	}
	switch {
	case e1.entry.Timestamp.UnixNano() < e2.entry.Timestamp.UnixNano():
		return true
	case e1.entry.Timestamp.UnixNano() > e2.entry.Timestamp.UnixNano():
		return false
	default:
		return e1.labels < e2.labels
	}
}

func mergeStreams(resps []*LokiResponse, limit uint32, direction logproto.Direction) []logproto.Stream {
	output := byDirection{
		direction: direction,
		entries:   []entry{},
	}
	for _, resp := range resps {
		for _, stream := range resp.Data.Result {
			for _, e := range stream.Entries {
				output.entries = append(output.entries, entry{
					entry:  e,
					labels: stream.Labels,
				})
			}
		}
	}
	sort.Sort(output)
	// limit result
	if len(output.entries) >= int(limit) {
		output.entries = output.entries[:limit]
	}

	resultDict := map[string]*logproto.Stream{}
	for _, e := range output.entries {
		stream, ok := resultDict[e.labels]
		if !ok {
			stream = &logproto.Stream{
				Labels:  e.labels,
				Entries: []logproto.Entry{},
			}
			resultDict[e.labels] = stream
		}
		stream.Entries = append(stream.Entries, e.entry)

	}
	keys := make([]string, 0, len(resultDict))
	for key := range resultDict {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	result := make([]logproto.Stream, 0, len(resultDict))
	for _, key := range keys {
		result = append(result, *resultDict[key])
	}

	return result
}
