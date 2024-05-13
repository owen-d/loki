package buffoon

import (
	"container/heap"
	"fmt"

	"go.uber.org/atomic"
)

/*
Blocks: The atom of buffer pool. Basically `[n]byte`
Pages: A contiguous memory range used to mount blocks. Basically `[n]byte`
Slabs: collection of pages that need not be contiguous blocks of memory. Basically `[][n]byte`.
Controller: The exposed management for the pkg. concurrency-safe
*/

type Addr struct {
	Slab   int
	Page   int
	Offset int
	Len    int
}

// TODO(owen-d): lots of improvements to be made if necessary.
// There's no need for actual slabs afaict, but it's easier to route
// via a couple heaps rather than a more complex defragmentation scheme.
type page struct {
	addr []byte

	// page contains two heaps with the same contents,
	// but with different ordering schemes:

	// min heap orders by address. Used to stitch contiguous ranges back
	// together after occupied blocks are released
	addresses *blockPQ

	// max heap ordered by the amount of contiguous memory. Used to provide
	// the next available block
	free *blockPQ

	occupied *atomic.Int64 // bytes occupied
}

func newPage(sz int) *page {
	p := &page{
		addr: make([]byte, sz),
		addresses: newBlockPriorityQueue(
			func(a, b *blockAddr) bool {
				return a.Offset < b.Offset
			},
		),
		free: newBlockPriorityQueue(
			func(a, b *blockAddr) bool {
				return a.Len > b.Len
			},
		),
	}

	entirety := &blockAddr{
		Len:    len(p.addr),
		Offset: 0,
	}

	heap.Push(p.addresses, entirety)
	heap.Push(p.free, entirety)

	return p
}

func (p *page) deFrag() {
	// TODO(owen-d): lock, iterate addresses, combine contiguous ranges,
	// and remove/update them from free.
	panic("unimplemented")
}

func (p *page) spaceFor(n int) bool {
	top := p.free.Peek()
	return top != nil && top.Len >= n
}

func (p *page) push(b []byte) *blockAddr {
	top := heap.Pop(p.free).(*blockAddr)
	if top.Len < len(b) {
		panic(
			fmt.Sprintf(
				"not enough space in page's highest contiguous range (len=%d) to add block (len=%d)",
				top.Len, len(b),
			),
		)
	}

	// split addr into two
	_ = p.occupied.Add(int64(len(b)))

	occupied := &blockAddr{
		Len:    len(b),
		Offset: top.Offset,
	}
	remaining := &blockAddr{
		Len:    top.Len - occupied.Len,
		Offset: top.Offset + occupied.Len,
	}

	heap.Push(p.free, remaining)
	return occupied
}

func (p *page) release(addr blockAddr) {
	panic("unimplemented")
}

func (p *page) ByteCapacity() int {
	return len(p.addr)
}

func (p *page) OccupiedPercentage() float64 {
	cap := p.ByteCapacity()
	if cap == 0 {
		return 0.
	}

	return float64(p.OccupiedBytes()) / float64(cap)
}

func (p *page) OccupiedBytes() int {
	panic("unimplemented")
}

func (p *page) NumKeys() int {
	panic("unimplemented")
}

type Slab struct {
	blockSize int
	pageSize  int
	pages     []page
}

func (s *Slab) ByteCapacity() int {
	return s.pageSize * len(s.pages)
}

type Controller struct {
	// TODO(owen-d): is this even necessary? answer: probably?
	// it helps preventing memory fragmentation i think.
	tiers []Slab
}
