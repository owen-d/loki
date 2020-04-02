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
	Labels = Braces(
		BindWith3(
			Characters,
			MatchType,
			Quotes(Characters),
			func(label interface{}, matchType interface{}, value interface{}) interface{} {
				return labels.MustNewMatcher(
					matchType.(labels.MatchType),
					label.(string),
					value.(string),
				)
			},
		),
	)
)
