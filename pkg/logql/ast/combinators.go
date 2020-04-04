package ast

import (
	"errors"
	"fmt"
)

// Unit constructs a ConstParser, which always succeeds with a specified value
// without consuming the input string.
func Unit(x interface{}) ConstParser {
	return ConstParser{
		t:   fmt.Sprintf("%T", x),
		val: x,
	}
}

// OneOf constructs a chain of AlternativeParsers: effectively a parser
// which attempts one of many parsers.
func OneOf(xs ...Parser) Parser {
	if len(xs) == 0 {
		return ErrParser{errors.New("No available options")}
	}

	res := xs[0]
	for _, x := range xs[1:] {
		res = Option(res, x)
	}
	return res

}

// Surround returns a parser that parses A, B, then C, but only returns the result from B
func Surround(a, b, c Parser) Parser {
	return BindWith3(a, b, c, func(_ interface{}, result interface{}, _ interface{}) interface{} {
		return result
	})
}

// BindWith2 is sugar for a 2 Parser Bind chain, allowing the implementor
// to specify a result with the results of the 2 previous parsers as arguments
func BindWith2(
	a, b Parser,
	fn func(interface{}, interface{}) interface{},
) MonadParser {
	return Bind(
		a,
		a.Type(),
		func(res1 interface{}) Parser {
			return Bind(
				b,
				b.Type(),
				func(res2 interface{}) Parser {
					return Unit(fn(res1, res2))
				},
			)
		},
	)
}

// BindWith3 is sugar for a 3 Parser Bind chain, allowing the implementor
// to specify a result with the results of the 3 previous parsers as arguments
func BindWith3(
	a, b, c Parser,
	fn func(interface{}, interface{}, interface{}) interface{},
) MonadParser {
	return Bind(
		a,
		a.Type(),
		func(res1 interface{}) Parser {
			return Bind(
				b,
				b.Type(),
				func(res2 interface{}) Parser {
					return Bind(
						c,
						c.Type(),
						func(res3 interface{}) Parser {
							return Unit(fn(res1, res2, res3))
						},
					)
				},
			)
		},
	)
}

// Satisfy returns a parser which is only successful if a given predicate passes.
// Otherwise it returns an ErrParser with the given error.
func Satisfy(predicate func(interface{}) bool, failureErr error, p Parser) MonadParser {
	return Bind(
		p,
		p.Type(),
		func(v interface{}) Parser {
			if predicate(v) {
				return Unit(v)
			}
			return ErrParser{failureErr}
		},
	)
}
