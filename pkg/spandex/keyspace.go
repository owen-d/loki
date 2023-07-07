package spandex

import (
	"fmt"
	"math"
)

type key uint64

func (k *key) String() string {
	if k == nil {
		return "nil"
	}
	return fmt.Sprintf("%d", uint64(*k))
}

func (k *key) Cmp(other *key) Cmp {
	if k == nil && other == nil {
		return Eq
	}

	if k == nil {
		return Gt
	}

	if other == nil {
		return Lt
	}

	if *k < *other {
		return Lt
	}

	if *k == *other {
		return Eq
	}

	return Gt
}

// Distance function, describes how far away two keys are
func (k key) Distance(other key) uint64 {
	if k > other {
		return uint64(k - other)
	}
	return uint64(other - k)
}

// A Keyspace is an arbitrary domain
// Modeled off of object storage lexicographic DHT.
type keyspace struct {
	From    key
	Through *key
}

func newKeySpace(from key, through *key) keyspace {
	return keyspace{
		From:    from,
		Through: through,
	}
}

// [From, through)
// through may be nil to indicate wrapping to end of keyspace
func (ks keyspace) Bounds() (from key, through *key) {
	return ks.From, ks.Through
}

// Owned tests if a key is owned by a keyspace.
func (ks keyspace) Owned(k key) bool {
	if ks.From.Cmp(&k) == Gt {
		return false
	}

	if ks.Through.Cmp(&k) != Gt {
		return false
	}

	return true
}

func (left keyspace) Intersect(right keyspace) keyspace {
	var newFrom key
	var newThrough *key

	switch left.From.Cmp(&right.From) {
	case Eq, Gt:
		newFrom = left.From
	case Lt:
		newFrom = right.From
	}

	switch left.Through.Cmp(right.Through) {
	case Lt, Eq:
		newThrough = left.Through
	case Gt:
		newThrough = right.Through
	}

	return keyspace{
		From:    newFrom,
		Through: newThrough,
	}

}

func (left keyspace) Union(right keyspace) keyspace {
	var newFrom key
	var newThrough *key

	switch left.From.Cmp(&right.From) {
	case Lt, Eq:
		newFrom = left.From
	case Gt:
		newFrom = right.From
	}

	switch left.Through.Cmp(right.Through) {
	case Gt:
		newThrough = left.Through
	case Lt, Eq:
		newThrough = right.Through
	}

	return keyspace{
		From:    newFrom,
		Through: newThrough,
	}
}

func (ks keyspace) Center() key {
	from, through := ks.Bounds()
	var max key
	if through == nil {
		max = math.MaxUint64
	} else {
		max = *through
	}

	diff := max - from
	return key(max - diff/2)
}
