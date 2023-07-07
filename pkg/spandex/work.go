package spandex

type work struct {
	name   string // identifier, i.e. "chunks_iterated" or "file_size"
	weight int    // allows specifying some forms of work as more important than others
	val    int
}

type workload []work

type worker struct {
	workload
	keyspace
}

// If a node owns a keyspace and is overloaded, where do we split it into two? Ideally in the middle of a hot spot.
// keyspace:
// A                   B       C                                  D
// [-------------------*********----------------------------------]
//    ^low_traffic^    ^hotspot^  ^low_traffic^
//
// Ideally we split this into the following:
// A                       B   C                                       D
// [-------------------****] | [*****----------------------------------]
//
// So we need to keep track of two pieces:
// (1) The total workload for the keyspace
// (2) The middle point of the owned keyspace at which half the work is on either side. Exactness is not required but finding a good compute_cost(memory|cpu) to accuracy ratio is important.
// TODO(research): https://en.wikipedia.org/wiki/Mainline_DHT -- something like k-buckets?

type node struct {
	keyspace
}
