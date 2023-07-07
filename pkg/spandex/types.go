package spandex

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
