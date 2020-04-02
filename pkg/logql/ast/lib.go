package ast

import (
	"strings"
)

var (
	LBracket = StringParser{"["}
	RBracket = StringParser{"]"}
	LBrace   = StringParser{"{"}
	RBrace   = StringParser{"}"}
	LParen   = StringParser{"("}
	RParen   = StringParser{")"}

	Equals = StringParser{"="}
	Add    = StringParser{"+"}
	Sub    = StringParser{"-"}
	Div    = StringParser{"/"}
	Mul    = StringParser{"*"}

	Bang = StringParser{"!"}

	CharSet = OneOfStrings(
		strings.Split(
			`/.*+?0123456789abcdedfghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ`,
			"",
		)...,
	)

	Characters = FMap(
		func(in interface{}) interface{} {
			return assertStr(in.([]interface{}))
		},
		ManyParser{CharSet},
		"Characters",
	)
)

func assertStr(in []interface{}) string {
	var out string
	for _, x := range in {
		out += x.(string)
	}
	return out
}

// Surround returns a parser that parses A, B, then C, but only returns the result from B
func Surround(a, b, c Parser) Parser {
	return Bind(
		a,
		a.Type(),
		func(_ interface{}) Parser {
			return Bind(
				b,
				b.Type(),
				func(result interface{}) Parser {
					return Bind(
						c,
						c.Type(),
						func(_ interface{}) Parser {
							return Unit(result)
						},
					)
				},
			)
		},
	)
}

// Parens is a parser combinator for wrapping a parser in parens
func Parens(p Parser) Parser {
	return Surround(LParen, p, RParen)
}

// Brackets is a parser combinator for wrapping a parser in brackets
func Brackets(p Parser) Parser {
	return Surround(LBracket, p, RBracket)
}

// Braces is a parser combinator for wrapping a parser in braces
func Braces(p Parser) Parser {
	return Surround(LBrace, p, RBrace)
}
