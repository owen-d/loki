package parser

import (
	"errors"
	"strconv"
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
		AssertStr,
		SomeParser{AlphaNumericParser},
		"AlphaNumerics",
	)

	Characters = FMap(
		AssertStr,
		SomeParser{CharSetParser},
		"Characters",
	)

	// utilities
	ManySpaces = ManyParser{StringParser{" "}}
	SomeSpaces = SomeParser{StringParser{" "}}

	IntParser = FMap(
		func(in interface{}) interface{} {
			s := AssertStr(in)
			n, _ := strconv.Atoi(s.(string))
			return n
		},
		SomeParser{OneOfStrings(strings.Split(Numeric, "")...)},
		"Integer",
	)
)

// AssertStr transforms an underlying slice of strings into a single string
func AssertStr(in interface{}) interface{} {
	var out string
	for _, x := range in.([]interface{}) {
		out += x.(string)
	}
	return out
}

func OneOfStrings(xs ...string) Parser {
	parsers := make([]Parser, 0, len(xs))
	for _, x := range xs {
		parsers = append(parsers, StringParser{x})
	}
	return OneOf(parsers...)

}

// Sequence simply returns a parser which
// composes a list of string parsers in order
// contract: The parsers must return strings.
func SequenceStrings(xs ...Parser) Parser {
	if len(xs) == 0 {
		return ErrParser{errors.New("No available parsers to sequence")}
	}

	result := xs[0]
	for _, p := range xs[1:] {
		result = ConcatStrings(result, p)
	}

	return result
}

// ConcatStrings returns a parser that concats two successive string parsers
func ConcatStrings(a, b Parser) MonadParser {
	return BindWith2(a, b, func(aRes, bRes interface{}) interface{} {
		return aRes.(string) + bRes.(string)
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

// WhiteSpaced returns a parser which trims whitespace around a given parser
func WhiteSpaced(p Parser) Parser {
	return Surround(ManySpaces, p, ManySpaces)
}
