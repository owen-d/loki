package logql

import (
	"github.com/grafana/loki/pkg/logql/parser"
	"github.com/prometheus/prometheus/pkg/labels"
)

var (
	MatchEqual = parser.FMap(
		parser.Const(labels.MatchEqual),
		parser.Equals,
		"labels.MatchEquals",
	)

	MatchNotEqual = parser.FMap(
		parser.Const(labels.MatchNotEqual),
		parser.StringP("!="),
		"labels.MatchNotEqual",
	)

	MatchRegexp = parser.FMap(
		parser.Const(labels.MatchRegexp),
		parser.StringP("=~"),
		"labels.MatchRegexp",
	)

	MatchNotRegexp = parser.FMap(
		parser.Const(labels.MatchNotRegexp),
		parser.StringP("!~"),
		"labels.MatchNotRegexp",
	)

	MatchType = parser.OneOf(MatchRegexp, MatchNotRegexp, MatchNotEqual, MatchEqual)

	// Labels is a parser for syntax such as {foo="bar"}
	Labels = parser.FMap(
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
	singleLabelParser = parser.BindWith3(
		parser.SomeAlphaNumerics,
		parser.WhiteSpaced(MatchType),
		parser.Quotes(parser.Characters),
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

	multiLabelParser = parser.Braces(
		parser.WhiteSpaced(
			parser.Separated(commaOptionalSpaces, singleLabelParser),
		),
	)

	// ,<  >
	commaOptionalSpaces = parser.SequenceStrings(
		parser.StringP(","),
		parser.FMap(
			parser.AssertStr,
			parser.ManySpaces,
			parser.ManySpaces.Type(),
		),
	)
)
