package spandex

import "fmt"

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

type keyspace struct {
	From    key
	Through *key
}

func (ks keyspace) Bounds() (from key, through *key) {
	return ks.From, ks.Through
}

func (ks keyspace) Owned(k key) bool {
	if ks.From.Cmp(&k) == Gt {
		return false
	}

	if ks.Through.Cmp(&k) != Gt {
		return false
	}

	return true
}

func (ks keyspace) Reduce(from key, through *key) keyspace {
	var newFrom key
	var newThrough *key

	switch ks.From.Cmp(&from) {
	case Eq, Gt:
		newFrom = ks.From
	case Lt:
		newFrom = from
	}

	switch ks.Through.Cmp(through) {
	case Lt, Eq:
		newThrough = ks.Through
	case Gt:
		newThrough = through
	}

	return keyspace{
		From:    newFrom,
		Through: newThrough,
	}

}

type file struct {
	Addr string
}

func (f file) Address() string {
	return f.Addr
}
