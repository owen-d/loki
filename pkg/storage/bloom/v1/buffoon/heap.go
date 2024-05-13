package buffoon

type blockAddr struct {
	Len, Offset int
	// index in the queue, used for re-establishing heap invariants in the
	// double-heap after mutation
	Index int
}

type blockPQ struct {
	addrs []*blockAddr
	less  func(a, b *blockAddr) bool
}

func newBlockPriorityQueue(
	less func(a, b *blockAddr) bool,
) *blockPQ {
	return &blockPQ{
		less: less,
	}
}

func (h blockPQ) Len() int { return len(h.addrs) }

func (h blockPQ) Less(i, j int) bool {
	// We want Pop to give us the highest, not lowest, priority so we use greater than here.
	return h.less(h.addrs[i], h.addrs[j])
}

func (h blockPQ) Swap(i, j int) {
	h.addrs[i], h.addrs[j] = h.addrs[j], h.addrs[i]
	h.addrs[i].Index = i
	h.addrs[j].Index = j
}

func (h *blockPQ) Push(x any) {
	item := x.(*blockAddr)
	item.Index = len(h.addrs)
	h.addrs = append(h.addrs, item)
}

func (h *blockPQ) Pop() any {
	old := h.addrs
	n := len(old)
	item := old[n-1]
	old[n-1] = nil // avoid memory leak
	item.Index = -1
	h.addrs = old[0 : n-1]
	return item
}

// Leaks
func (h *blockPQ) Peek() *blockAddr {
	if len(h.addrs) > 0 {
		return (h.addrs)[0]
	}
	return nil
}

func (h *blockPQ) update(item *blockAddr)
