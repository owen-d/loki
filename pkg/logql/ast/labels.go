package ast

import (
	"github.com/prometheus/prometheus/pkg/labels"
)

var (
	MatchEqual = FMap(
		Const(labels.MatchEqual),
		Equals,
		"labels.MatchEquals",
	)

	MatchNotEqual = FMap(
		Const(labels.MatchNotEqual),
		StringParser{"!="},
		"labels.MatchNotEqual",
	)

	MatchRegexp = FMap(
		Const(labels.MatchRegexp),
		StringParser{"=~"},
		"labels.MatchRegexp",
	)

	MatchNotRegexp = FMap(
		Const(labels.MatchNotRegexp),
		StringParser{"!~"},
		"labels.MatchNotRegexp",
	)

	MatchType = OneOf(MatchRegexp, MatchNotRegexp, MatchNotEqual, MatchEqual)

	// Labels is a parser for syntax such as {foo="bar"}
	Labels = FMap(
		func(xs interface{}) interface{} {
			var res []*labels.Matcher

			for _, x := range xs.([]interface{}) {
				res = append(res, x.(*labels.Matcher))
			}
			return res
		},
		multiLabelParser,
		"[]*labels.Matcher",
	)
)

// non exported parsers, generally used for composition by exported forms
var (
	singleLabelParser = BindWith3(
		SomeAlphaNumerics,
		Surround(ManySpaces, MatchType, ManySpaces),
		Quotes(Characters),
		func(
			label interface{},
			matchType interface{},
			value interface{},
		) interface{} {
			return labels.MustNewMatcher(
				matchType.(labels.MatchType),
				label.(string),
				value.(string),
			)
		},
	)

	multiLabelParser = Braces(
		Surround(
			ManySpaces,
			Separated(
				commaOptionalSpaces,
				singleLabelParser,
			),
			ManySpaces,
		),
	)

	// ,<  >
	commaOptionalSpaces = SequenceStrings(
		StringParser{","},
		FMap(
			assertStr,
			ManySpaces,
			ManySpaces.Type(),
		),
	)
)

// Separated extracts a separated list. Requires at least 1 entry
// like: `a, a,a` -> `[]{a,a,a}`
func Separated(sepBy, p Parser) Parser {
	return BindWith2(
		ManyParser{First(p, sepBy)},
		p,
		func(aRes, bRes interface{}) interface{} {
			xs := aRes.([]interface{})
			return append(xs, bRes)
		},
	)
}

// matches a parser 0 or 1 times
func Maybe(p Parser) Parser {
	return Option(p, Unit(""))
}

// First sequences a two parser chain, discarding the second result
// (<<)
func First(a, b Parser) Parser {
	return BindWith2(a, b, func(aRes, bRes interface{}) interface{} {
		return aRes
	})
}

// Second sequences a two parser chain, discarding the first result
// (>>)
func Second(a, b Parser) Parser {
	return BindWith2(a, b, func(aRes, bRes interface{}) interface{} {
		return bRes
	})
}
