package planning

import (
	"fmt"
)

type LiteralColumn struct {
	value any
	n     int
	dtype DataTypeSignal
}

func (l *LiteralColumn) N() int {
	return l.n
}

func (l *LiteralColumn) Type() DataTypeSignal {
	return l.dtype
}

func (l *LiteralColumn) At(i int) (any, error) {
	if i < 0 || i >= l.n {
		return nil, fmt.Errorf("index out of range")
	}
	return l.value, nil
}

// SliceColumn represents a column storing items in a slice
type SliceColumn struct {
	items []any
	dtype DataTypeSignal
}

func (s *SliceColumn) N() int {
	return len(s.items)
}

func (s *SliceColumn) Type() DataTypeSignal {
	return s.dtype
}

func (s *SliceColumn) At(i int) (any, error) {
	if i < 0 || i >= len(s.items) {
		return nil, fmt.Errorf("index out of range")
	}
	return s.items[i], nil
}
