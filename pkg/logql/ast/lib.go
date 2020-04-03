package ast

import (
	"strings"
)

var (
	// character lists
	Alpha          = "abcdedfghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	Numeric        = "0123456789"
	ExtraRESymbols = "/.*+?!" // non-exhaustive
	AlphaNumeric   = Alpha + Numeric
	CharSet        = AlphaNumeric + ExtraRESymbols

	LBracket  = StringParser{"["}
	RBracket  = StringParser{"]"}
	LBrace    = StringParser{"{"}
	RBrace    = StringParser{"}"}
	LParen    = StringParser{"("}
	RParen    = StringParser{")"}
	Quotation = StringParser{`"`}

	Equals = StringParser{"="}
	Add    = StringParser{"+"}
	Sub    = StringParser{"-"}
	Div    = StringParser{"/"}
	Mul    = StringParser{"*"}

	Bang = StringParser{"!"}

	CharSetParser = OneOfStrings(
		strings.Split(CharSet, "")...,
	)

	AlphaNumericParser = OneOfStrings(
		strings.Split(AlphaNumeric, "")...,
	)

	SomeAlphaNumerics = FMap(
		assertStr,
		SomeParser{AlphaNumericParser},
		"AlphaNumerics",
	)

	Characters = FMap(
		assertStr,
		SomeParser{CharSetParser},
		"Characters",
	)

	// utilities
	ManySpaces = ManyParser{StringParser{" "}}
	SomeSpaces = SomeParser{StringParser{" "}}
)

func assertStr(in interface{}) interface{} {
	var out string
	for _, x := range in.([]interface{}) {
		out += x.(string)
	}
	return out
}

// Surround returns a parser that parses A, B, then C, but only returns the result from B
func Surround(a, b, c Parser) Parser {
	return BindWith3(a, b, c, func(_ interface{}, result interface{}, _ interface{}) interface{} {
		return result
	})
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

// Braces is a parser combinator for wrapping a parser in braces
func Quotes(p Parser) Parser {
	return Surround(Quotation, p, Quotation)
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
