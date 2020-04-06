package logql

import (
	"github.com/grafana/loki/pkg/logql/parser"
	"github.com/prometheus/prometheus/pkg/labels"
)

var (
	FilterEqual = parser.FMap(
		parser.Const(labels.MatchEqual),
		parser.StringP("|="),
		"FilterEqual",
	)

	FilterNotEqual = parser.FMap(
		parser.Const(labels.MatchNotEqual),
		parser.StringP("!="),
		"FilterNotEqual",
	)

	FilterRegexp = parser.FMap(
		parser.Const(labels.MatchRegexp),
		parser.StringP("|~"),
		"FilterRegexp",
	)

	FilterNotRegexp = parser.FMap(
		parser.Const(labels.MatchNotRegexp),
		parser.StringP("!~"),
		"FilterNotRegexp",
	)

	FilterType = parser.OneOf(FilterEqual, FilterNotEqual, FilterRegexp, FilterNotRegexp)

	// whitespaced<filtertype + space + quotes<somealphanum>>
	FilterP = parser.WhiteSpaced(parser.BindWith3(
		FilterType,
		parser.ManySpaces,
		parser.Quotes(parser.SomeAlphaNumerics),
		func(fType, _, str interface{}) interface{} {
			return Filter{
				ty:    fType.(labels.MatchType),
				match: str.(string),
			}
		},
	))

	FiltersP = parser.FMap(
		func(xs interface{}) interface{} {
			var res []Filter
			for _, x := range xs.([]interface{}) {
				res = append(res, x.(Filter))
			}
			return res
		},
		parser.ManyP(parser.WhiteSpaced(FilterP)),
		"[]Filters",
	)
)

type Filter struct {
	ty    labels.MatchType
	match string
}
