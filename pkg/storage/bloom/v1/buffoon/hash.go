package buffoon

import "sync"

type Hashable[T comparable] interface {
	Hash() T
}

type H32 interface {
	Hashable[uint32]
}

// TODO
type Striped[K H32, V any] struct {
	n       int
	locks   []sync.RWMutex
	buckets []V
}
