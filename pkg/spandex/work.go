package spandex

import "math"

// If a node owns a keyspace and is overloaded, where do we split it into two? Ideally in the middle of a hot spot.
// keyspace:
// A                  B center  C                                  D
// [------------------*****|*****---------------------------------]
//    ^low_traffic^    ^hotspot^  ^low_traffic^
//
// Ideally we split this into the following:
// A                       B   C                                       D
// [------------------*****] | [******---------------------------------]
//
// So we need to keep track of two pieces:
// (1) The total workload for the keyspace
// (2) The middle point of the owned keyspace at which half the work is on either side. Exactness is not required but finding a good compute_cost(memory|cpu) to accuracy ratio is important.
// TODO(research): https://en.wikipedia.org/wiki/Mainline_DHT -- something like k-buckets?

const (
	bucketBitFactor = 4
)

type node struct {
	address key // the address of this node
	// a bucket for each 4 bits in the uint64 to keep track of distances
	// roughly this means a bucket to track work for keys
	// within the first `64/16 == 4` bits away
	// This equates to keys with distances of `2^4=16` within the address of the node
	// The same is true for the next 4 bits -- `2^8 = 256`, creating a bucket
	// for keys with distances between the current & previous bucket:
	// `16 <= diff < 256`
	// This means we keep track of key accesses with exponentially decaying granularity based on their distance from the node's address.
	buckets [64 / bucketBitFactor]bucket
}

func (n *node) Pressure() (res int) {
	for _, b := range n.buckets {
		res += b.left
		res += b.right
	}
	return
}

// function to decay pressure over time
func (n *node) Decay() {
	factor := 50 // percent decay, example
	for i := range n.buckets {
		left, right := n.buckets[i].left, n.buckets[i].right
		left = left * factor / 100
		right = right * factor / 100
		n.buckets[i].left = left
		n.buckets[i].right = right
	}
}

func (n *node) Record(k key, weight int) {
	dist := n.address.Distance(k)
	position := bucketFor(dist)

	// technically we weight the node's address itself on the left side
	// which is incorrect but rare enough and close enough
	// to not make a significant difference (i think)
	if n.address.Cmp(&k) != Lt {
		n.buckets[position].left += weight
	} else {
		n.buckets[position].right += weight
	}
}

func bucketFor(distance uint64) int {
	bits := uint64(math.Log2(float64(distance)))
	return int(bits / bucketBitFactor)
}

type bucket struct {
	// Left & right track some measurement of work done in a bucket
	// of some given distance range from a node's address
	// Since buckets are stored based on distance from the node's addr,
	// left & right denote whether the accessed key was lesser or greater
	// than the node
	left, right int
}
