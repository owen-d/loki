package spandex

// A Keyspace is an arbitrary domain
// Modeled off of object storage lexicographic DHT.
type KeySpace interface {
	// [From, through)
	// through may be nil to indicate wrapping to end of keyspace
	Bounds() (from Key, through Key)

	// Owned tests if a key is owned by a keyspace.
	Owned(Key) bool

	// Reduce creates a subset keyspace from an existing one.
	// through may be nil to indicate wrapping to end of keyspace
	Reduce(from Key, through Key) KeySpace
}

// why the fuck am i writing go
type Cmp uint8

const (
	Eq Cmp = iota
	Lt
	Gt
)

func (c Cmp) String() string {
	switch c {
	case Eq:
		return "Eq"
	case Lt:
		return "Lt"
	case Gt:
		return "Gt"
	default:
		panic("undefined comparison value")
	}
}

// A Key is an associated type of a keyspace. TODO: make generic?
type Key interface {
	Cmp(Key) Cmp
}

// A Resource is something owned
type Resource interface {
	Address() Key
}

// Work tracks a measure of work done. requests, cpu cycles, data iterated, file sizes, etc
type Work interface{}
